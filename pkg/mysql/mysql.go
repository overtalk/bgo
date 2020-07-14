package mysqlpkg

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"github.com/overtalk/bgo/pkg/log"
	"github.com/overtalk/bgo/utils/xml"
)

const (
	ASC  OrderT = "ASC"
	DESC OrderT = "DESC"
)

type (
	OrderT string
	Order  struct {
		T      OrderT
		Fields []string
	}
	Condition struct {
		Op    string
		Value interface{}
	}
	RowHandler  = func(row *sql.Row) error
	RowsHandler = func(row *sql.Rows) error
)

type Config struct {
	XMLName  xml.Name `xml:"xml"`
	PoolSize int      `xml:"poolsize"`
	Username string   `xml:"username"`
	Password string   `xml:"password"`
	Address  string   `xml:"address"`
	Dbname   string   `xml:"dbname"`
}

type MysqlConn struct {
	cfg *Config
	db  *sql.DB
}

func NewMysqlConn(path string) (*MysqlConn, error) {
	cfg := &Config{}
	if err := xmlutil.ParseXml(path, cfg); err != nil {
		return nil, err
	}

	return &MysqlConn{cfg: cfg}, nil
}

func (this *MysqlConn) Connect() error {
	cfg := &mysql.Config{
		User:                 this.cfg.Username,
		Passwd:               this.cfg.Password,
		Addr:                 this.cfg.Address,
		DBName:               this.cfg.Dbname,
		Loc:                  time.Now().Location(),
		ParseTime:            true,
		Net:                  "tcp",
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		//logpkg.GetLogger().With(zap.Error(err)).Error("failed to open mysql")
		logpkg.Error("failed to open mysql", zap.Error(err))
		return err
	}
	db.SetMaxOpenConns(this.cfg.PoolSize)

	if maxIdle := this.cfg.PoolSize / 10; maxIdle > 2 {
		db.SetMaxIdleConns(maxIdle)
	}

	if err = db.Ping(); err != nil {
		logpkg.Error("failed to ping mysql", zap.Error(err))
		return err
	}

	this.db = db
	return nil
}

func (this *MysqlConn) GetConn() *sql.DB { return this.db }

func (this *MysqlConn) Insert(tableName string, data map[string]interface{}) (int64, error) {
	sql, values := GenInsertSql(tableName, data)
	res, err := this.db.Exec(sql, values...)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (this *MysqlConn) Update(tableName string, data map[string]interface{}, where map[string]Condition) (int64, error) {
	sql, values := GenUpdateSql(tableName, data, where)
	res, err := this.db.Exec(sql, values...)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (this *MysqlConn) SelectOne(tableName string, columns []string, where map[string]Condition, order *Order) (map[string]interface{}, error) {
	res, err := this.Select(tableName, columns, where, order, 1)
	if err != nil {
		return nil, err
	}

	if len(res) < 1 {
		return nil, errors.New("no rows found")
	}

	return res[0], nil
}

func (this *MysqlConn) Select(tableName string, columns []string, where map[string]Condition, order *Order, limit int) ([]map[string]interface{}, error) {
	var ret []map[string]interface{}
	rowsHandler := func(rows *sql.Rows) error {
		rowColumns, err := rows.Columns()
		if err != nil {
			return err
		}
		rowColumnCount := len(rowColumns)

		for rows.Next() {
			scanFrom := make([]interface{}, rowColumnCount)
			scanTo := make([]interface{}, rowColumnCount)
			for i, _ := range scanFrom {
				scanFrom[i] = &scanTo[i]
			}

			if err := rows.Scan(scanFrom...); err != nil {
				return err
			}

			temp := make(map[string]interface{})
			for i, _ := range scanTo {
				temp[rowColumns[i]] = scanTo[i]
			}

			ret = append(ret, temp)
		}
		return nil
	}

	return ret, this.SelectWithHandler(tableName, columns, where, order, limit, rowsHandler)
}

func (this *MysqlConn) SelectOneWithHandler(tableName string, columns []string, where map[string]Condition, order *Order, handler RowHandler) error {
	sql, values := GenSelectSql(tableName, columns, where, order, 1)
	row := this.db.QueryRow(sql, values...)
	return handler(row)
}

func (this *MysqlConn) SelectWithHandler(tableName string, columns []string, where map[string]Condition, order *Order, limit int, handler RowsHandler) error {
	sql, values := GenSelectSql(tableName, columns, where, order, limit)
	rows, err := this.db.Query(sql, values...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return handler(rows)
}

// generate sql
func GenSelectSql(tableName string, columns []string, where map[string]Condition, order *Order, limit int) (string, []interface{}) {
	sql := fmt.Sprintf("SELECT %s FROM `%s`", strings.Join(columns, ","), tableName)

	var values []interface{}
	if len(where) > 0 {
		var whereKeys []string
		for k, v := range where {
			if v.Value != nil {
				whereKeys = append(whereKeys, fmt.Sprintf(" `%s`%s?", k, v.Op))
				values = append(values, v.Value)
			} else {
				// for special where condition, eg : IS NOT NULL
				whereKeys = append(whereKeys, fmt.Sprintf(" `%s`%s", k, v.Op))
			}
		}

		sql = fmt.Sprintf("%s WHERE %s", sql, strings.Join(whereKeys, ","))
	}

	if order != nil {
		sql = fmt.Sprintf("%s ORDER BY %s %s", sql, strings.Join(order.Fields, ","), string(order.T))
	}

	if limit > 0 {
		sql = fmt.Sprintf("%s LIMIT %d", sql, limit)
	}

	//logpkg.GetLogger().With(zap.String("sql", sql), zap.Any("values", values)).Debug("mysql select")
	logpkg.Debug("mysql select", zap.String("sql", sql), zap.Any("values", values))

	return sql, values
}

func GenInsertSql(tableName string, data map[string]interface{}) (string, []interface{}) {
	var keys []string
	var placeholders []string
	var values []interface{}
	for k, v := range data {
		keys = append(keys, k)
		placeholders = append(placeholders, "?")
		values = append(values, v)
	}

	// generate sql
	sql := fmt.Sprintf("INSERT into `%s` (%s) value (%s)", tableName, strings.Join(keys, ","), strings.Join(placeholders, ","))

	//logpkg.GetLogger().With(zap.String("sql", sql), zap.Any("values", values)).Debug("mysql insert")
	logpkg.Debug("mysql insert", zap.String("sql", sql), zap.Any("values", values))

	return sql, values
}

func GenUpdateSql(tableName string, data map[string]interface{}, where map[string]Condition) (string, []interface{}) {
	var keys []string
	var values []interface{}
	for k, v := range data {
		keys = append(keys, fmt.Sprintf(" `%s`=?", k))
		values = append(values, v)
	}

	sql := fmt.Sprintf("UPDATE `%s` set %s", tableName, strings.Join(keys, ","))

	if len(where) > 0 {
		var whereKeys []string
		for k, v := range where {
			if v.Value != nil {
				whereKeys = append(whereKeys, fmt.Sprintf(" `%s`%s?", k, v.Op))
				values = append(values, v.Value)
			} else {
				// for special where condition, eg : IS NOT NULL
				whereKeys = append(whereKeys, fmt.Sprintf(" `%s`%s", k, v.Op))
			}
		}

		sql = fmt.Sprintf("%s where %s", sql, strings.Join(whereKeys, ","))
	}

	//logpkg.GetLogger().With(zap.String("sql", sql), zap.Any("values", values)).Debug("mysql update")
	logpkg.Debug("mysql update", zap.String("sql", sql), zap.Any("values", values))

	return sql, values
}
