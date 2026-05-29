package pagination

import (
	"go-echo-demo/internal/constants"
)

// 查询场景标识符 由前端传递
type SceneID string

// 查询场景参数
type SceneParams map[string]any

// 构建查询
type SceneBuilder func(params SceneParams) (PageQuery, error)

// 查询场景注册表
type Registry struct {
	builders map[SceneID]SceneBuilder
}

func NewRegistry() *Registry {
	return &Registry{
		builders: make(map[SceneID]SceneBuilder),
	}
}

func (r *Registry) Register(id SceneID, builder SceneBuilder) {
	r.builders[id] = builder
}

func (r *Registry) Build(id SceneID, params SceneParams) (PageQuery, error) {
	builder, ok := r.builders[id]
	if !ok {
		return PageQuery{}, constants.UnknownScene
	}
	return builder(params)
}

// KnownScenes 返回所有已注册的 scene ID，用于文档生成或调试接口
func (r *Registry) KnownScenes() []SceneID {
	ids := make([]SceneID, 0, len(r.builders))
	for id := range r.builders {
		ids = append(ids, id)
	}
	return ids
}
