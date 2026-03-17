package research

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ppt-smasher/internal/config"

	"github.com/google/uuid"
)

// ParseWithMinerU 调用 MinerU 在线服务解析 PDF 并解压结果
func ParseWithMinerU(ctx context.Context, filePath string) (string, []string, error) {
	mineruCfg := config.GlobalConfig.MinerU
	if mineruCfg.APIKey == "" {
		return "", nil, fmt.Errorf("MinerU API Key is not set")
	}

	// 1. 获取上传链接 (POST /file-urls/batch)
	uploadURLReq := fmt.Sprintf("%s/file-urls/batch", mineruCfg.BaseURL)
	fileName := filepath.Base(filePath)

	payload := map[string]interface{}{
		"files": []map[string]string{
			{"name": fileName, "data_id": uuid.New().String()},
		},
		"model_version": "vlm",
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURLReq, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	fmt.Printf("MinerU API Key: %s", mineruCfg.APIKey)
	req.Header.Set("Authorization", "Bearer "+mineruCfg.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("failed to get upload URL (%d): %s", resp.StatusCode, string(resBody))
	}

	var uploadResp struct {
		Code int `json:"code"`
		Data struct {
			BatchID  string   `json:"batch_id"`
			FileURLs []string `json:"file_urls"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", nil, err
	}
	if uploadResp.Code != 0 || len(uploadResp.Data.FileURLs) == 0 {
		return "", nil, fmt.Errorf("MinerU API error: %s (code: %d)", uploadResp.Msg, uploadResp.Code)
	}

	// 2. 上传文件 (PUT)
	file, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	uploadReq, err := http.NewRequestWithContext(ctx, "PUT", uploadResp.Data.FileURLs[0], file)
	if err != nil {
		return "", nil, err
	}
	// 文档提示上传无须设置 Content-Type

	uploadRespRaw, err := client.Do(uploadReq)
	if err != nil {
		return "", nil, err
	}
	defer uploadRespRaw.Body.Close()
	if uploadRespRaw.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to upload file: status %d", uploadRespRaw.StatusCode)
	}

	// 3. 轮询结果 (GET /extract-results/batch/{batch_id})
	resultURL := fmt.Sprintf("%s/extract-results/batch/%s", mineruCfg.BaseURL, uploadResp.Data.BatchID)
	for i := 0; i < 60; i++ { // 最多 10 分钟 (60 * 10s)
		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		case <-time.After(10 * time.Second):
			req, _ = http.NewRequestWithContext(ctx, "GET", resultURL, nil)
			req.Header.Set("Authorization", "Bearer "+mineruCfg.APIKey)

			resp, err = client.Do(req)
			if err != nil {
				continue
			}

			var res struct {
				Code int `json:"code"`
				Data struct {
					ExtractResults []struct {
						State      string `json:"state"`
						FullZipURL string `json:"full_zip_url"`
						ErrMsg     string `json:"err_msg"`
					} `json:"extract_result"`
				} `json:"data"`
			}
			json.NewDecoder(resp.Body).Decode(&res)
			resp.Body.Close()

			if len(res.Data.ExtractResults) > 0 {
				item := res.Data.ExtractResults[0]
				if item.State == "done" {
					// 4. 下载并解压
					outputDir := config.GlobalConfig.Paths.MinerUResult
					if outputDir == "" {
						outputDir = "mineru_result"
					}
					err := downloadAndExtract(ctx, item.FullZipURL, outputDir)
					if err != nil {
						return "", nil, fmt.Errorf("failed to download and extract MinerU result: %v", err)
					}
					return fmt.Sprintf("MinerU extract successful. Results extracted to %s/", outputDir), nil, nil
				}
				if item.State == "failed" {
					return "", nil, fmt.Errorf("MinerU extraction failed: %s", item.ErrMsg)
				}
			}
		}
	}

	return "", nil, fmt.Errorf("MinerU timeout")
}

func downloadAndExtract(ctx context.Context, zipURL string, outputDir string) error {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// 下载 ZIP 文件
	req, err := http.NewRequestWithContext(ctx, "GET", zipURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download zip: status %d", resp.StatusCode)
	}

	// 读取到内存
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 解压
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		path := filepath.Join(outputDir, f.Name)

		// 检查路径安全
		if !filepath.HasPrefix(path, filepath.Clean(outputDir)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		srcFile, err := f.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
