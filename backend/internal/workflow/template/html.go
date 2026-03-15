package template

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino/compose"
	"github.com/unidoc/unioffice/presentation"
	"github.com/unidoc/unioffice/schema/soo/dml"
)

func getShapeCSS(spPr *dml.CT_ShapeProperties) string {
	if spPr == nil || spPr.Xfrm == nil {
		return ""
	}

	xfrm := spPr.Xfrm
	var topPt, leftPt, widthPt, heightPt float64

	if xfrm.Off != nil {
		if xfrm.Off.XAttr.ST_CoordinateUnqualified != nil {
			leftEMU := *xfrm.Off.XAttr.ST_CoordinateUnqualified
			leftPt = float64(leftEMU) / 12700.0
		}
		if xfrm.Off.YAttr.ST_CoordinateUnqualified != nil {
			topEMU := *xfrm.Off.YAttr.ST_CoordinateUnqualified
			topPt = float64(topEMU) / 12700.0
		}
	}

	if xfrm.Ext != nil {
		widthEMU := xfrm.Ext.CxAttr
		heightEMU := xfrm.Ext.CyAttr
		widthPt = float64(widthEMU) / 12700.0
		heightPt = float64(heightEMU) / 12700.0
	}

	return fmt.Sprintf(`style="position: absolute; top: %.2fpt; left: %.2fpt; width: %.2fpt; height: %.2fpt;"`,
		topPt, leftPt, widthPt, heightPt)
}

func layoutToHTML(pptPath string) ([]SlideHTMLSchema, error) {
	if _, err := os.Stat(pptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("reference ppt not found: %s", pptPath)
	}

	pres, err := presentation.Open(pptPath)
	if err != nil {
		return nil, err
	}
	defer pres.Close()

	ss := pres.SlideSize()
	slideSize := ss.X()
	widthPt := float64(slideSize.CxAttr) / 12700.0
	heightPt := float64(slideSize.CyAttr) / 12700.0

	var htmlViews []SlideHTMLSchema

	for _, layout := range pres.SlideLayouts() {
		html := fmt.Sprintf("<!DOCTYPE html>\n<html>\n<body style=\"width:%.2fpt; height:%.2fpt;\">\n", widthPt, heightPt)

		if layout.X() != nil && layout.X().CSld != nil && layout.X().CSld.SpTree != nil {
			for _, choice := range layout.X().CSld.SpTree.Choice {
				for _, shape := range choice.Sp {
					css := getShapeCSS(shape.SpPr)
					text := "Text Placeholder"
					html += fmt.Sprintf("  <div %s>\n    <p>%s</p>\n  </div>\n", css, text)
				}
				for _, pic := range choice.Pic {
					css := getShapeCSS(pic.SpPr)
					html += fmt.Sprintf("  <img alt=\"Image Placeholder\" %s>\n", css)
				}
			}
		}

		html += "</body>\n</html>\n"

		htmlViews = append(htmlViews, SlideHTMLSchema{
			LayoutName: layout.Name(),
			HTML:       html,
		})
	}

	return htmlViews, nil
}

func NewHTMLRendererNode() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, s TeamTemplateState) (TeamTemplateState, error) {
		log.Println("[TemplateAnalyst] 正在渲染为 HTML 视图导航......")

		if s.ReferencePPT != "" {
			views, err := layoutToHTML(s.ReferencePPT)
			if err != nil {
				log.Printf("[TemplateAnalyst] HTML 转换失败: %v", err)
			} else {
				s.HTMLViews = views
				log.Printf("[TemplateAnalyst] 成功将 %d 种布局转为 HTML", len(views))
			}
		}

		return s, nil
	})
}
