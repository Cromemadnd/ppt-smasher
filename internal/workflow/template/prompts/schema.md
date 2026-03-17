请分析以下幻灯片版式的 HTML 结构，并创建一个涵盖幻灯片元素的 JSON Schema。

## 任务要求：
1. 理解提供的 HTML 结构，特别是其中的样式、布局以及元素间的相互关联关系。
2. 为 HTML 中的每个元素（`<div>` 中的文本或 `<img>`）提取以下信息：
    - "name": 该元素的作用和特征描述。例如 "section/main/sub title"（章节/主/副标题）, "left/right bullets"（左右列表）, "portrait/landscape/square image"（垂直/水平/方形图片）, "slide/section number"（幻灯片/分类序号）等。请确保名称描述能够准确代表其在幻灯片上的位置和用途。
    - "type": "text" 或 "image"。
    - "id": 根据提供的信息中提取对应的 Placeholder ID。如果没有提供，可忽略。
    - "index": 根据提供的信息中提取对应的 Placeholder Index。如果没有提供，可忽略。

## 输出格式要求：
严格输出包含 "elements" 数组的 JSON。
不要输出除 JSON 之外的任何 Markdown 和解释性文本。

输入版式 HTML：
{{html}}