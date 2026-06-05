// playground/firestore/main.go
//
// Firestore API 练习场
// 这不是标准测试文件，直接 `go run ./playground/firestore` 运行
//
// 使用方式：
//   1. 在 main() 里按需取消注释你想练习的函数
//   2. go run ./playground/firestore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// ──────────────────────────────────────────────
// 凭证 & 初始化
// ──────────────────────────────────────────────

const credFile = "/Users/lin/go-study/go-echo-demo/go-echo-demo-firebase-adminsdk-fbsvc-98c4d740c5.json"

// 练习用的临时集合，不污染正式数据
const playCollection = "playground_tasks"

func initClient(ctx context.Context) *firestore.Client {
	opt := option.WithCredentialsFile(credFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("初始化 Firebase App 失败: %v", err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("初始化 Firestore Client 失败: %v", err)
	}
	return client
}

// ──────────────────────────────────────────────
// 数据结构（练习专用，不复用 dto）
// ──────────────────────────────────────────────

type PlayTask struct {
	Title       string    `firestore:"title"`
	Description string    `firestore:"description"`
	Status      int       `firestore:"status"` // 0=todo 1=in_progress 2=done
	Priority    int       `firestore:"priority"`
	Tags        []string  `firestore:"tags"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
}

// ──────────────────────────────────────────────
// main：取消注释你想跑的练习
// ──────────────────────────────────────────────

func main() {
	ctx := context.Background()
	client := initClient(ctx)
	defer client.Close()

	fmt.Println("=== Firestore 练习场已启动 ===")

	// --- 基础 CRUD ---
	// practiceAdd(ctx, client)
	// practiceSet(ctx, client)
	// practiceGet(ctx, client, "task-manual-001")
	// practiceUpdate(ctx, client, "task-manual-001")
	// practiceDelete(ctx, client, "task-manual-001")

	// --- 批量操作 ---
	// practiceBatchWrite(ctx, client)
	// practiceBulkUpsert(ctx, client)
	// practiceBulkGet(ctx, client, []string{"<id1>", "<id2>"})

	// --- 查询 ---
	// practiceQuery(ctx, client)
	// practiceQueryOrderAndLimit(ctx, client)
	// practiceQueryComposite(ctx, client)
	// practiceQueryArray(ctx, client)

	// --- 实时监听 ---
	// practiceListenDocument(ctx, client, "123456")
	// practiceListenQuery(ctx, client)

	// --- 事务 ---
	practiceTransaction(ctx, client, "123456")

	// --- 清理练习数据 ---
	// cleanPlayground(ctx, client)

	fmt.Println("完成，请在 main() 里取消注释你想练习的函数")
}

// ══════════════════════════════════════════════
// 基础 CRUD
// ══════════════════════════════════════════════

// practiceAdd 自动生成 ID 写入文档
func practiceAdd(ctx context.Context, client *firestore.Client) {
	fmt.Println("\n── Add（自动生成ID）──")

	task := PlayTask{
		Title:       "learn Firestore Go SDK",
		Description: "learn deeply firestore go sdk,include  create update read delete,and transaction controll",
		Status:      1,
		Priority:    1,
		Tags:        []string{"学习", "firestore", "go"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add method this method will automatically generate doc id
	docRef, _, err := client.Collection(playCollection).Add(ctx, task)
	if err != nil {
		return
	}

	fmt.Printf("your task id is %s", docRef.ID)
	fmt.Printf("the playgrouond task created %v", docRef)
}

// practiceSet 使用指定 ID 写入（覆盖整个文档）
func practiceSet(ctx context.Context, client *firestore.Client) {
	fmt.Println("\n── Set（指定ID，全量覆盖）──")

	docID := "task-manual-001"
	task := PlayTask{
		Title:       "读 Firestore 文档",
		Description: "把官方文档过一遍",
		Status:      1,
		Priority:    1,
		Tags:        []string{"学习"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	dataMap, err := ToMap(task)
	if err != nil {
		log.Print("map converter error", err)
		return
	}

	// if docRef.ID exists,this action will update the existed doc
	// if not exists , this action will insert a new doc ,but not report error
	_, err = client.Collection(playCollection).Doc(docID).Set(ctx,
		dataMap, firestore.MergeAll)
	if err != nil {
		log.Print("Set Doc error", err)
		return
	}
	fmt.Printf("Set 成功，文档 ID: %s\n", docID)
}

// practiceGet 读取单个文档
func practiceGet(ctx context.Context, client *firestore.Client, docID string) {
	fmt.Printf("\n── Get（读取文档 %s）──\n", docID)

	docSnap, err := client.Collection(playCollection).Doc(docID).Get(ctx)
	if err != nil {
		log.Print("get doc err docId:", docID)
		return
	}
	log.Printf("you got the docSnap is %v", docSnap)

	var res PlayTask
	// DataTo method will map the docSnap to Struct
	err = docSnap.DataTo(&res)
	if err != nil {
		log.Print("DataTo error")
		return
	}
	log.Printf("you got the struct result is %v by docId:%s", res, docID)

	// Data method will map the docSnap to map[string]interface{}
	m := docSnap.Data()
	log.Printf("you got the map result is %v by docId:%s", m, docID)

	// DataAt will get the spec field's value
	// if the field that you provide not exist , it will be report error no field ...
	title, err := docSnap.DataAt("title")
	if err != nil {
		log.Print("DataAt error", err)
		return
	}
	log.Printf("you got the sepc filed value is %v", title)

	// the docSnap.Ref.Id will return the docRef Id
	id := docSnap.Ref.ID
	log.Printf("you got the docId is %s", id)

	// the docSnap.Ref.Path will return the complete path of the docSnap
	path := docSnap.Ref.Path
	log.Printf("the docSnap path is %s", path)
}

// practiceUpdate 局部更新（只改指定字段，不覆盖整个文档）
func practiceUpdate(ctx context.Context, client *firestore.Client, docID string) {
	fmt.Printf("\n── Update（局部更新 %s）──\n", docID)

	// 方式一：Update + []firestore.Update（推荐，类型安全）
	// this method will report error if the doc not exist
	// the difference with Set+MergeAll is the doc if exist has a difference fluence
	_, err := client.Collection(playCollection).Doc(docID).Update(ctx, []firestore.Update{
		{
			Path:  "status",
			Value: 2,
		},
		{
			Path:  "title",
			Value: "update the title",
		},
		{
			Path:  "unknown_field",
			Value: "this is unknown field value",
		},
	})
	if err != nil {
		log.Printf("update the doc failed %v", err)
		return
	}
	log.Printf("update the %s doc succeed", docID)
	// 方式二：Set + MergeAll（用 map，灵活）
	// this method not check the doc's existence
	// _, err = client.Collection(playCollection).Doc(docID).Set(ctx, map[string]any{
	// 	"priority":   3,
	// 	"updated_at": time.Now(),
	// }, firestore.MergeAll)
}

// practiceDelete 删除文档
func practiceDelete(ctx context.Context, client *firestore.Client, docID string) {
	fmt.Printf("\n── Delete（删除文档 %s）──\n", docID)

	// delete the entire doc
	// _, err := client.Collection(playCollection).Doc(docID).Delete(ctx)
	// if err != nil {
	// 	log.Printf("delete the doc that docId is %s failed", docID)
	// 	return
	// }

	// delete the single field
	_, err := client.Collection(playCollection).Doc(docID).Update(ctx, []firestore.Update{
		{Path: "description", Value: firestore.Delete},
	})
	if err != nil {
		log.Printf("delete the single field description failed")
		return
	}
	log.Printf("delete the description field is success")
}

// ══════════════════════════════════════════════
// 批量操作
// ══════════════════════════════════════════════

// practiceBatchWrite 原子批量写（不能跨文档读，只写）
func practiceBatchWrite(ctx context.Context, client *firestore.Client) {
	fmt.Println("\n── BulkWriter / WriteBatch（批量写）──")

	batch := client.Batch()

	ids := []string{
		"batch-task-A", "batch-task-B", "batch-task-C",
	}
	tasks := []PlayTask{
		{Title: "批量任务A", Status: 0, Priority: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "批量任务B", Status: 0, Priority: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "批量任务C", Status: 1, Priority: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for i, t := range tasks {
		taskRef := client.Collection(playCollection).Doc(ids[i])
		batch.Set(taskRef, t)
	}

	_, err := batch.Commit(ctx)
	if err != nil {
		log.Printf("the batch write is failed %v", err)
	}
	log.Printf("the batch write is succeed")
}

// 通过bulk writer批量写入
func practiceBulkUpsert(ctx context.Context, client *firestore.Client) {
	bw := client.BulkWriter(ctx)

	ids := []string{
		"bulk-task-A", "bulk-task-B", "bulk-task-C", "bulk-task-D", "bulk-task-E",
	}

	tasks := []PlayTask{
		{Title: "批量任务A", Status: 0, Priority: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "批量任务B", Status: 0, Priority: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "批量任务C", Status: 1, Priority: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "批量任务D", Status: 1, Priority: 4, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Title: "批量任务E", Status: 1, Priority: 4, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	jobs := make([]*firestore.BulkWriterJob, 0, len(tasks))

	for i, task := range tasks {
		docRef := client.Collection(playCollection).Doc(ids[i])
		data, err := ToMap(task)
		if err != nil {
			bw.End()
			log.Printf("the bulk write is quit")
		}
		// 先删除 再写入
		jobAdd, err := bw.Set(docRef, data, firestore.MergeAll)
		jobs = append(jobs, jobAdd)
	}
	// 发送所有已排队的写入，并关闭 BulkWriter
	bw.End()

	// 逐个检查写入结果
	for i, job := range jobs {
		if _, err := job.Results(); err != nil {
			log.Printf("write task %s failed: %v", ids[i], err)
		}
	}

}

// practiceBulkGet 通过多个 ID 批量读取（事务外）
func practiceBulkGet(ctx context.Context, client *firestore.Client, ids []string) {
	fmt.Println("\n── GetAll（批量读取）──")

	refs := make([]*firestore.DocumentRef, len(ids))
	for i, id := range ids {
		refs[i] = client.Collection(playCollection).Doc(id)
	}

	snaps, err := client.GetAll(ctx, refs)
	if err != nil {
		log.Printf("GetAll 失败: %v", err)
		return
	}

	for _, snap := range snaps {
		if !snap.Exists() {
			fmt.Printf("文档 %s 不存在\n", snap.Ref.ID)
			continue
		}
		var task PlayTask
		_ = snap.DataTo(&task)
		fmt.Printf("ID: %s, Title: %s, Status: %d\n", snap.Ref.ID, task.Title, task.Status)
	}

}

// ══════════════════════════════════════════════
// 查询
// ══════════════════════════════════════════════

// practiceQuery 基础 Where 查询
func practiceQuery(ctx context.Context, client *firestore.Client) {
	fmt.Println("\n── Query（Where 过滤）──")

	// iter := client.Collection(playCollection).Where("status", "==", 0).Where("title", "==", "批量任务A").Limit(10).Documents(ctx)
	iter := client.Collection(playCollection).Where("tags", "array-contains", "学习").Limit(10).Documents(ctx)

	for {
		// 迭代读取
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}
		res := doc.Data()
		log.Printf("the current doc is %s \n", res)
	}
}

// practiceQueryOrderAndLimit 排序 + 分页（Limit / Offset）
func practiceQueryOrderAndLimit(ctx context.Context, client *firestore.Client) {
	fmt.Println("\n── Query（OrderBy + Limit）──")

	iter := client.Collection(playCollection).
		OrderBy("priority", firestore.Desc).
		Limit(3).
		Documents(ctx)
	defer iter.Stop()

	for {
		snap, err := iter.Next()
		if err != nil {
			break
		}
		var task PlayTask
		_ = snap.DataTo(&task)
		fmt.Printf("Priority: %d  Title: %s\n", task.Priority, task.Title)
	}

	// ── Cursor 分页示例 ──
	// 先拿第一页最后一条：
	// lastSnap := ...
	// 下一页：
	// iter2 := client.Collection(playCollection).
	// 	OrderBy("priority", firestore.Desc).
	// 	StartAfter(lastSnap).
	// 	Limit(3).
	// 	Documents(ctx)
}

// practiceQueryComposite 复合条件查询（需要在 Firestore 控制台创建复合索引）
func practiceQueryComposite(ctx context.Context, client *firestore.Client) {
	fmt.Println("\n── Query（复合条件）──")

	iter := client.Collection(playCollection).
		Where("status", "==", 0).
		Where("priority", ">=", 2).
		OrderBy("priority", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()

	for {
		snap, err := iter.Next()
		if err != nil {
			break
		}
		var task PlayTask
		_ = snap.DataTo(&task)
		fmt.Printf("ID: %s  Priority: %d  Title: %s\n", snap.Ref.ID, task.Priority, task.Title)
	}
}

// practiceQueryArray 数组字段查询：array-contains
func practiceQueryArray(ctx context.Context, client *firestore.Client) {
	fmt.Println("\n── Query（array-contains）──")

	iter := client.Collection(playCollection).
		Where("tags", "array-contains", "学习").
		Documents(ctx)
	defer iter.Stop()

	for {
		snap, err := iter.Next()
		if err != nil {
			break
		}
		var task PlayTask
		_ = snap.DataTo(&task)
		fmt.Printf("ID: %s  Tags: %v  Title: %s\n", snap.Ref.ID, task.Tags, task.Title)
	}
}

// ══════════════════════════════════════════════
// 实时监听
// ══════════════════════════════════════════════

// practiceListenDocument 监听单个文档变更（阻塞 10 秒）
func practiceListenDocument(ctx context.Context, client *firestore.Client, docID string) {
	fmt.Printf("\n── Listen（监听文档 %s，10秒）──\n", docID)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	iter := client.Collection(playCollection).Doc(docID).Snapshots(ctx)
	defer iter.Stop()

	for {
		snap, err := iter.Next()
		if err != nil {
			fmt.Printf("监听结束: %v\n", err)
			return
		}
		if !snap.Exists() {
			fmt.Println("文档不存在或已被删除")
			continue
		}
		var task PlayTask
		_ = snap.DataTo(&task)
		fmt.Printf("文档变更: %+v\n", task)
	}
}

// practiceListenQuery 监听查询结果集变更（阻塞 10 秒）
func practiceListenQuery(ctx context.Context, client *firestore.Client) {
	fmt.Println("\n── Listen（监听查询，10秒）──")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	iter := client.Collection(playCollection).
		Where("status", "==", 0).
		Snapshots(ctx)
	defer iter.Stop()

	for {
		snap, err := iter.Next()
		if err != nil {
			fmt.Printf("监听结束: %v\n", err)
			return
		}
		fmt.Printf("当前查询结果共 %d 条，变更 %d 条\n", snap.Size, len(snap.Changes))
		for _, change := range snap.Changes {
			fmt.Printf("  变更类型: %v  ID: %s\n", change.Kind, change.Doc.Ref.ID)
		}
	}
}

// ══════════════════════════════════════════════
// 事务
// ══════════════════════════════════════════════

// practiceTransaction 演示事务：读后写（计数器递增）
func practiceTransaction(ctx context.Context, client *firestore.Client, docID string) {
	fmt.Printf("\n── Transaction（文档 %s 的 priority +1）──\n", docID)

	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		log.Printf("事务启动...\n")
		// 事务内读
		snap, err := tx.Get(client.Collection(playCollection).Doc(docID))
		if err != nil {
			return err
		}

		var task PlayTask
		if err := snap.DataTo(&task); err != nil {
			return err
		}

		docRef := client.Collection(playCollection).Doc("task-manual-001")
		err = tx.Update(docRef, []firestore.Update{
			{Path: "title", Value: "测试事务内部修改"},
			{Path: "priority", Value: 2},
		})

		// 手动伪造一个错误 观察是否会修改成功
		// err = errors.New("this is a manual")
		// if err != nil {
		// 	return err
		// }

		// 休眠十秒 从外部修改数据 看看是否会重试
		time.Sleep(15 * time.Second)

		// 基于读到的值做修改
		return tx.Update(snap.Ref, []firestore.Update{
			{Path: "priority", Value: task.Priority + 1},
			{Path: "updated_at", Value: time.Now()},
		})
	})

	if err != nil {
		log.Printf("Transaction 失败: %v", err)
		return
	}
	fmt.Println("Transaction 成功")
}

// ══════════════════════════════════════════════
// 清理
// ══════════════════════════════════════════════

// cleanPlayground 删除 playCollection 下所有练习数据
func cleanPlayground(ctx context.Context, client *firestore.Client) {
	fmt.Printf("\n── 清理集合 %s ──\n", playCollection)

	iter := client.Collection(playCollection).Documents(ctx)
	defer iter.Stop()

	batch := client.Batch()
	count := 0

	for {
		snap, err := iter.Next()
		if err != nil {
			break
		}
		batch.Delete(snap.Ref)
		count++
	}

	if count == 0 {
		fmt.Println("集合已为空，无需清理")
		return
	}

	if _, err := batch.Commit(ctx); err != nil {
		log.Printf("清理失败: %v", err)
		return
	}
	fmt.Printf("已删除 %d 条练习数据\n", count)
}
