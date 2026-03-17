package ws

import (
	"log"

	"github.com/gorilla/websocket"
)

// Client 代理单个 WebSocket 连接
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan Message
}

// SendMsg 实现跨包 Sender 接口
func (c *Client) SendMsg(msgType string, payload interface{}) {
	c.Send <- Message{
		Type:    msgType,
		Payload: payload,
	}
}

// ReadPump 抽水机：持续读取客户端发来的消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// 路由处理消息
		HandleClientMessage(c, message)
	}
}

// WritePump 送水机：将后端数据写入 WebSocket 发送到客户端
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteJSON(msg)
		}
	}
}
