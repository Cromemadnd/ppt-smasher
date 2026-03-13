import { useState, useEffect, useCallback, useRef } from 'react';

// WebSocket message types (matching backend)
export const MsgTypeStartTask = 'START_TASK';
export const MsgTypeUploadFile = 'UPLOAD_FILE';
export const MsgTypeUpdateFileDesc = 'UPDATE_FILE_DESC';

export const MsgTypeKnowledgeBaseUpdate = 'KNOWLEDGE_BASE_UPDATE';
export const MsgTypeNodeActive = 'NODE_ACTIVE';
export const MsgTypeEdgeActive = 'EDGE_ACTIVE';
export const MsgTypeAgentThoughtStream = 'AGENT_THOUGHT_STREAM';
export const MsgTypeStepCompleted = 'STEP_COMPLETED';
export const MsgTypeTaskSuccess = 'TASK_SUCCESS';
export const MsgTypeTaskError = 'TASK_ERROR';

export interface WSMessage {
    type: string;
    payload: any;
}

export function useWebSocket(url: string) {
    const [messages, setMessages] = useState<WSMessage[]>([]);
    const ws = useRef<WebSocket | null>(null);

    useEffect(() => {
        ws.current = new WebSocket(url);

        ws.current.onopen = () => {
            console.log('WebSocket connected');
        };

        ws.current.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                console.log('Received:', msg);
                setMessages((prev) => [...prev, msg]);
            } catch (err) {
                console.error('Failed to parse WebSocket message:', event.data);
            }
        };

        ws.current.onclose = () => {
            console.log('WebSocket disconnected');
        };

        return () => {
            if (ws.current) {
                ws.current.close();
            }
        };
    }, [url]);

    const sendMessage = useCallback((type: string, payload: any = {}) => {
        if (ws.current && ws.current.readyState === WebSocket.OPEN) {
            const msg: WSMessage = { type, payload };
            ws.current.send(JSON.stringify(msg));
        } else {
            console.error('WebSocket is not open');
        }
    }, []);

    // Return specific latest messages handlers or raw messages
    return { sendMessage, messages, ws: ws.current };
}
