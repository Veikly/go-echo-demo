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
	taskGroup.GET("", server.TaskHandler.ListTasks)          // 分页查询
	taskGroup.POST("/archive", server.TaskHandler.BatchArchieveTask) // 批量归档（事务）

	userGroup := apiGroup.Group("/users")
	userGroup.GET("/me", server.UserHandler.GetMyDetail) // 查询当前登录用户，ID 从 token 中取
	userGroup.POST("", server.UserHandler.CompleteUserInfo)

}
