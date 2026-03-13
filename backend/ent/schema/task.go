package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.String("theme").NotEmpty().Comment("任务主题"),
		field.String("status").Default("pending").Comment("任务状态: pending, generating, success, error"),
		field.Text("result_url").Optional().Comment("最终产物链接"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}
