package db

import (
	"context"
	"database/sql"
)

// DBTX 是 sql.DB 和 sql.Tx 的接口.
type DBTX interface {
	// ExecContext 执行一个查询但不返回任何行.
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	// PrepareContext 为以后的查询或执行创建一个预处理语句.
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	// QueryContext 执行一个返回行的查询，通常是 SELECT.
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	// QueryRowContext 执行一个预期最多返回一行的查询.
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type Queries struct {
	db                          DBTX      // 数据库连接
	tx                          *sql.Tx   // 事务
	createFileStmt              *sql.Stmt // 创建文件
	createMessageStmt           *sql.Stmt // 创建消息
	createSessionStmt           *sql.Stmt // 创建会话
	deleteFileStmt              *sql.Stmt // 删除文件
	deleteMessageStmt           *sql.Stmt // 删除消息
	deleteSessionStmt           *sql.Stmt // 删除会话
	deleteSessionFilesStmt      *sql.Stmt // 删除会话文件
	deleteSessionMessagesStmt   *sql.Stmt // 删除会话消息
	getFileStmt                 *sql.Stmt // 获取文件
	getFileByPathAndSessionStmt *sql.Stmt // 获取文件
	getMessageStmt              *sql.Stmt // 获取消息
	getSessionByIDStmt          *sql.Stmt // 获取会话
	listFilesByPathStmt         *sql.Stmt // 列出文件
	listFilesBySessionStmt      *sql.Stmt // 列出文件
	listLatestSessionFilesStmt  *sql.Stmt // 列出最新文件
	listMessagesBySessionStmt   *sql.Stmt // 列出消息
	listNewFilesStmt            *sql.Stmt // 列出新文件
	listSessionsStmt            *sql.Stmt // 列出会话
	updateMessageStmt           *sql.Stmt // 更新消息
	updateSessionStmt           *sql.Stmt // 更新会话
}
