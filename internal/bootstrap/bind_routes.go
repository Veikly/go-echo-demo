package bootstrap

// BindRoutes 为Echo实例绑定路由
func BindRoutes(server *Server) {
	apiGroup := server.Echo.Group("/api")

	// 定义TaskGroup
	taskGroup := apiGroup.Group("/tasks")
	taskGroup.POST("", server.TaskHandler.CreateTask)
	taskGroup.GET("/:id", server.TaskHandler.GetTaskDetail)
	taskGroup.PUT("/:id", server.TaskHandler.ModifyTask)
	taskGroup.DELETE("/:id", server.TaskHandler.DeleteTask)
	taskGroup.GET("", server.TaskHandler.ListTasks) // 分页查询

	userGroup := apiGroup.Group("/users")
	userGroup.GET("/:id", server.UserHandler.GetMyDetail)
	userGroup.POST("", server.UserHandler.CompleteUserInfo)

}
