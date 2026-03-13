package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// KnowledgeFile holds the schema definition for the KnowledgeFile entity.
type KnowledgeFile struct {
	ent.Schema
}

// Fields of the KnowledgeFile.
func (KnowledgeFile) Fields() []ent.Field {
	return []ent.Field{
		field.String("file_name").NotEmpty().Comment("文件名"),
		field.String("file_type").NotEmpty().Comment("文件类型 (pdf, docx, txt)"),
		field.Text("description").Optional().Comment("大模型生成的摘要或用户手动修改的描述"),
		field.String("storage_path").NotEmpty().Comment("本地存储全路径"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}
