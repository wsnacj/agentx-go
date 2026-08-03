# OCR Worker Pool API（Experimental）

`document/ocr/worker`提供基于semaphore的轻量并发限制器。

```go
pool := worker.NewPool(config.WorkerConfig{MaxConcurrent: 4, QueueSize: 8})
if err := pool.Acquire(ctx); err != nil { /* canceled or deadline */ }
defer pool.Release()
```

`Acquire`在有容量时立即成功，等待时传播context；`Release`对未持有slot的调用保持无阻塞，
但调用方仍应保证每次成功Acquire恰好Release一次。本包不创建goroutine、不拥有任务队列
持久化，也不是系统scheduler。`Pool`可并发使用，当前为Experimental。
