import React, { useState, useEffect } from 'react';
import {
    ReactFlow,
    Background,
    Controls,
    type Node,
    type Edge,
    MarkerType,
    Position,
    Handle,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import {
    FileText,
    Image as ImageIcon,
    MessageSquare,
    Search,
    Users,
    Brain,
    Database,
    PenTool,
    Monitor,
    LayoutTemplate
} from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import {
    useWebSocket,
    MsgTypeStartTask,
    MsgTypeKnowledgeBaseUpdate,
    MsgTypeNodeActive,
    MsgTypeEdgeActive,
    MsgTypeAgentThoughtStream,
} from './hooks/useWebSocket';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

// ----------------------------------------
// Types & Data
// ----------------------------------------
interface CustomNodeData {
    label: string;
    role?: string;
    icon?: React.ElementType;
    isActive?: boolean;
    status?: 'waiting' | 'running' | 'done';
}

interface FileItem {
    id: string;
    name: string;
    type: 'doc' | 'image';
    desc: string;
}

const mockFiles: FileItem[] = [
    { id: '1', name: '2023_Industry_Report.pdf', type: 'doc', desc: 'Industry overview and key stats' },
    { id: '2', name: 'Q4_Financials.xlsx', type: 'doc', desc: 'Raw financial data for Q4' },
    { id: '3', name: 'Company_Logo.png', type: 'image', desc: 'Brand logo for slides' },
];

const MOCK_ACTIVE_EDGES: string[] = [];

// ----------------------------------------
// Custom Nodes
// ----------------------------------------
const AgentNode = ({ data }: { data: CustomNodeData; id: string }) => {
    const Icon = (data.icon as React.ElementType) || Brain;
    const isActive = data.isActive;

    return (
        <div
            className={cn(
                "relative rounded-xl border-2 bg-white px-5 py-4 min-w-[200px] shadow-sm transition-all duration-300",
                isActive ? "border-blue-500 shadow-[0_0_15px_rgba(59,130,246,0.5)] scale-105" : "border-slate-200",
                "flex flex-col items-center gap-2"
            )}
        >
            <Handle type="target" position={Position.Top} className="!bg-blue-400 !w-3 !h-3" />
            <div className={cn("p-3 rounded-full", isActive ? "bg-blue-100 text-blue-600" : "bg-slate-100 text-slate-500")}>
                <Icon size={24} />
            </div>
            <div className="text-center">
                <h3 className="font-semibold text-slate-800">{data.label}</h3>
                {data.role && <p className="text-xs text-slate-500">{data.role}</p>}
            </div>
            <Handle type="source" position={Position.Bottom} className="!bg-blue-400 !w-3 !h-3" />
        </div>
    );
};

const DataNode = ({ data }: { data: CustomNodeData; id: string }) => {
    const Icon = (data.icon as React.ElementType) || Database;
    return (
        <div className="rounded-xl border-2 border-slate-200 bg-slate-50 px-6 py-4 flex items-center gap-3 shadow-sm min-w-[180px]">
            <Handle type="target" position={Position.Top} className="!bg-blue-400" />
            <div className="p-2 bg-white rounded-lg border border-slate-200 text-blue-500">
                <Icon size={20} />
            </div>
            <span className="font-medium text-slate-700">{data.label}</span>
            <Handle type="source" position={Position.Bottom} className="!bg-blue-400" />
        </div>
    );
};

const nodeTypes = {
    agent: AgentNode,
    data: DataNode,
};

// ----------------------------------------
// Tree Layout & Graph Setup
// ----------------------------------------
const initialNodes: Node[] = [
    // Level 1: User
    { id: 'user', type: 'agent', position: { x: 400, y: 50 }, data: { label: 'User', icon: Users } },

    // Level 2: Boss
    { id: 'boss', type: 'agent', position: { x: 400, y: 200 }, data: { label: 'Boss Agent', role: 'Core Coordinator', icon: Brain, isActive: false } },

    // Level 3: Researchers & Leaders
    { id: 'researcher', type: 'agent', position: { x: 100, y: 350 }, data: { label: 'Researcher', role: 'Data Gatherer', icon: Search } },
    { id: 'content', type: 'agent', position: { x: 400, y: 350 }, data: { label: 'Content Leader', role: 'Outline & Text', icon: PenTool } },
    { id: 'visual', type: 'agent', position: { x: 700, y: 350 }, data: { label: 'Visual Leader', role: 'Design & Layout', icon: LayoutTemplate } },

    // Level 4: Output / Data
    { id: 'vdb', type: 'data', position: { x: 250, y: 550 }, data: { label: 'Vector DataBase', icon: Database } },
    { id: 'ppt', type: 'data', position: { x: 550, y: 550 }, data: { label: 'Final PPT', icon: Monitor } },
];

const defaultEdgeOptions = {
    style: { strokeWidth: 2, stroke: '#94a3b8' },
    markerEnd: {
        type: MarkerType.ArrowClosed,
        color: '#94a3b8',
    },
};

const activeEdgeOptions = {
    style: { strokeWidth: 3, stroke: '#3b82f6' },
    animated: true,
    markerEnd: {
        type: MarkerType.ArrowClosed,
        color: '#3b82f6',
    },
};

const createEdge = (id: string, source: string, target: string, label: string) => ({
    id,
    source,
    target,
    label,
    labelBgPadding: [8, 4] as [number, number],
    labelBgBorderRadius: 4,
    labelBgStyle: { fill: '#fff', color: '#fff', fillOpacity: 0.9 },
    labelStyle: { fill: '#334155', fontWeight: 500, fontSize: 12 },
    ...(MOCK_ACTIVE_EDGES.includes(id) ? activeEdgeOptions : defaultEdgeOptions)
});

const initialEdges: Edge[] = [
    createEdge('user-boss', 'user', 'boss', ""),
    createEdge('boss-researcher', 'boss', 'researcher', ""),
    createEdge('researcher-boss', 'researcher', 'boss', ""),
    createEdge('boss-content', 'boss', 'content', ""),
    createEdge('content-boss', 'content', 'boss', ""),
    createEdge('boss-visual', 'boss', 'visual', ""),
    createEdge('researcher-vdb', 'researcher', 'vdb', ""),

    // Content Leader to VDB retrieve
    {
        id: 'content-vdb', source: 'content', target: 'vdb', label: 'Retrieve',
        ...defaultEdgeOptions,
        markerEnd: { type: MarkerType.ArrowClosed, color: '#94a3b8' },
        markerStart: { type: MarkerType.ArrowClosed, color: '#94a3b8' }, // Fake bidirectional visually, but react flow doesn't perfectly support markerStart arrows facing back well without custom edge, keeping simple or use 2 edges
    },

    createEdge('visual-ppt', 'visual', 'ppt', 'Generate PPT'),
];


export default function WorkflowUI() {
    const [files, setFiles] = useState<FileItem[]>(mockFiles);
    const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
    const [nodes, setNodes] = useState<Node[]>(initialNodes);
    const [edges, setEdges] = useState<Edge[]>(initialEdges);
    const [inputValue, setInputValue] = useState("");
    const [agentThoughts, setAgentThoughts] = useState<Record<string, string>>({});

    const { sendMessage, messages } = useWebSocket("ws://localhost:8080/ws");

    useEffect(() => {
        if (messages.length === 0) return;
        const lastMsg = messages[messages.length - 1];
        const { type, payload } = lastMsg;

        if (type === MsgTypeNodeActive) {
            setNodes((nds) => nds.map((n) => ({
                ...n,
                data: {
                    ...n.data,
                    isActive: n.id === payload.nodeId || payload.nodeId === 'all'
                }
            })));
        } else if (type === MsgTypeEdgeActive) {
            setEdges((eds) => eds.map((e) => ({
                ...e,
                ...(e.id === payload.edgeId
                    ? { ...activeEdgeOptions }
                    : { ...defaultEdgeOptions }
                )
            })));
        } else if (type === MsgTypeAgentThoughtStream) {
            setAgentThoughts((prev) => ({
                ...prev,
                [payload.nodeId]: (prev[payload.nodeId] || "") + payload.chunk
            }));
        } else if (type === MsgTypeKnowledgeBaseUpdate) {
            setFiles((prev) => [
                ...prev,
                { id: String(Date.now()), name: payload.fileName, type: 'doc', desc: payload.desc }
            ]);
        }
    }, [messages]);

    const handleSendTask = () => {
        if (!inputValue.trim()) return;
        sendMessage(MsgTypeStartTask, { theme: inputValue });
        setInputValue("");
        setAgentThoughts({});
    };

    const handleDescChange = (id: string, newDesc: string) => {
        setFiles(files.map(f => f.id === id ? { ...f, desc: newDesc } : f));
    };

    const selectedNode = nodes.find(n => n.id === selectedNodeId);

    return (
        <div className="flex h-screen w-full bg-slate-50 font-sans overflow-hidden">

            {/* ---------------- Sidebar (Knowledge Base) ---------------- */}
            <div className="w-[320px] bg-white border-r border-slate-200 flex flex-col shadow-sm z-10 flex-shrink-0">
                <div className="p-5 border-b border-slate-100">
                    <h2 className="text-xl font-bold text-slate-800 flex items-center gap-2">
                        <Database className="text-blue-500" size={24} />
                        知识库
                    </h2>
                    <button className="mt-4 w-full border-2 border-dashed border-blue-300 bg-blue-50 hover:bg-blue-100 text-blue-600 rounded-lg py-3 flex items-center justify-center gap-2 transition-colors font-medium">
                        + 上传文件...
                    </button>
                </div>

                <div className="flex-1 overflow-y-auto p-4 space-y-4">
                    {files.map(file => (
                        <div key={file.id} className="p-3 bg-slate-50 rounded-xl border border-slate-200 hover:border-blue-300 transition-colors group">
                            <div className="flex items-center gap-3 mb-2">
                                <div className="p-2 bg-white rounded-lg text-blue-500 shadow-sm">
                                    {file.type === 'doc' ? <FileText size={18} /> : <ImageIcon size={18} />}
                                </div>
                                <span className="font-medium text-slate-700 truncate text-sm flex-1">{file.name}</span>
                            </div>
                            <input
                                type="text"
                                value={file.desc}
                                onChange={(e) => handleDescChange(file.id, e.target.value)}
                                className="w-full text-xs bg-white border border-slate-200 rounded px-2 py-1.5 focus:outline-none focus:border-blue-400 focus:ring-1 focus:ring-blue-400 text-slate-600 transition-all"
                                placeholder="添加描述..."
                            />
                        </div>
                    ))}
                </div>
            </div>

            {/* ---------------- Main Area ---------------- */}
            <div className="flex-1 relative flex flex-col">

                {/* Top Input Bar */}
                <div className="absolute top-6 left-1/2 -translate-x-1/2 z-20 w-[600px] max-w-[90%] pointer-events-none">
                    <div className="bg-white/90 backdrop-blur-md border border-slate-200 shadow-lg rounded-2xl p-2 flex items-center gap-2 pointer-events-auto ring-1 ring-blue-500/10">
                        <input
                            type="text"
                            value={inputValue}
                            onChange={(e) => setInputValue(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && handleSendTask()}
                            placeholder="输入 PPT 生成主题、你的观点..."
                            className="flex-1 bg-transparent border-none outline-none px-4 py-2 text-slate-800 placeholder-slate-400"
                        />
                        <button
                            onClick={handleSendTask}
                            className="bg-blue-500 hover:bg-blue-600 text-white px-5 py-2.5 rounded-xl font-medium flex items-center gap-2 transition-colors shadow-sm shadow-blue-500/30 text-sm"
                        >
                            <MessageSquare size={16} />
                            发送
                        </button>
                    </div>
                </div>

                {/* React Flow Canvas */}
                <div className="flex-1">
                    <ReactFlow
                        nodes={nodes}
                        edges={edges}
                        nodeTypes={nodeTypes}
                        onNodeClick={(_, node) => setSelectedNodeId(node.id)}
                        onPaneClick={() => setSelectedNodeId(null)}
                        fitView
                        fitViewOptions={{ padding: 0.2 }}
                        minZoom={0.5}
                        maxZoom={1.5}
                    >
                        <Background color="#cbd5e1" gap={16} />
                        <Controls className="!bg-white !border-slate-200 !shadow-md rounded-lg overflow-hidden" />
                    </ReactFlow>
                </div>

                {/* ---------------- Thought Process Overlay (Right Slide-out) ---------------- */}
                <div className={cn(
                    "absolute top-4 bottom-4 right-4 w-[380px] bg-white rounded-2xl shadow-2xl border border-slate-200 transform transition-transform duration-300 ease-in-out flex flex-col overflow-hidden z-30",
                    selectedNode && selectedNode.type === 'agent' ? "translate-x-0" : "translate-x-[120%]"
                )}>
                    {selectedNode && selectedNode.type === 'agent' && (
                        <>
                            <div className="p-5 border-b border-slate-100 bg-gradient-to-r from-blue-50 to-white flex items-center gap-4">
                                <div className="p-3 bg-blue-500 text-white rounded-xl shadow-md shadow-blue-500/20">
                                    {selectedNode.data.icon ? (React.createElement(selectedNode.data.icon as React.ElementType, { size: 24 }) as React.ReactNode) : null}
                                </div>
                                <div>
                                    <h3 className="font-bold text-lg text-slate-800">{(selectedNode.data.label as React.ReactNode) || ""}</h3>
                                    <p className="text-sm text-blue-600 font-medium">{(selectedNode.data.role as React.ReactNode) || ""}</p>
                                </div>
                            </div>

                            <div className="flex-1 overflow-y-auto p-5">
                                {/* Thought Bubble */}
                                {agentThoughts[selectedNode.id] && (
                                    <div className="relative bg-blue-50 border border-blue-100 rounded-2xl rounded-tl-sm p-4 text-sm text-slate-700 italic mb-8 whitespace-pre-wrap flex flex-col gap-2">
                                        <div className="flex gap-2 mb-2 items-center text-blue-500 font-semibold border-b border-blue-100 pb-2">
                                            <Brain size={16} /> 思考过程
                                        </div>
                                        {agentThoughts[selectedNode.id]}
                                    </div>
                                )}

                                {/* Timeline */}
                                <h4 className="font-semibold text-slate-800 mb-4 px-1">执行流程与思考</h4>
                                <div className="space-y-6 relative before:absolute before:inset-0 before:ml-[11px] before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-0.5 before:bg-slate-200 pl-8">

                                    {/* Step 1: Done */}
                                    <div className="relative">
                                        <div className="absolute left-[-32px] w-5 h-5 bg-green-500 rounded-full border-4 border-white shadow-sm ring-1 ring-slate-200 flex items-center justify-center"></div>
                                        <h5 className="font-medium text-slate-800 text-sm">解析需求</h5>
                                        <p className="text-xs text-slate-500 mt-1">成功提取主题和核心观点</p>
                                    </div>

                                    {/* Step 2: Running */}
                                    <div className="relative">
                                        <div className="absolute left-[-32px] w-5 h-5 bg-blue-500 rounded-full border-4 border-white shadow-[0_0_8px_rgba(59,130,246,0.6)] animate-pulse flex items-center justify-center z-10"></div>
                                        <h5 className="font-medium text-blue-600 text-sm">检索知识库</h5>
                                        <p className="text-xs text-slate-600 mt-1">检索 "2023_Industry_Report.pdf" 中相关内容...</p>
                                    </div>

                                    {/* Step 3: Waiting */}
                                    <div className="relative">
                                        <div className="absolute left-[-32px] w-5 h-5 bg-slate-300 rounded-full border-4 border-white shadow-sm ring-1 ring-slate-200 flex items-center justify-center z-10"></div>
                                        <h5 className="font-medium text-slate-500 text-sm">生成大纲</h5>
                                        <p className="text-xs text-slate-400 mt-1">等待资料检索完成</p>
                                    </div>

                                </div>
                            </div>
                        </>
                    )}
                </div>

            </div>
        </div>
    );
}