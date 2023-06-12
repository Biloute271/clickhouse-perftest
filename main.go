package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	stanlog "github.com/Biloute271/stan-log"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gin-gonic/gin"
)

var config Config
var results []int64

func main() {
	stanlog.SetLogLevelDebug()
	err := readConfig()
	if err != nil {
		stanlog.Critical("Impossible to read configuration file / devices URL file")
		log.Fatal(err)
	}

	// insertRecords(1000, "nas")
	// insertRecords(1000, "s3_main")
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/:policy/:recordscount", launchBench)
	r.GET("/batch/:policy/:recordscount/:iterations", launchBatch)
	stanlog.Info("Starting Webserver")
	stanlog.Debug("Gin mode : " + gin.Mode())
	r.Run("0.0.0.0:8080") // 0.0.0.0 instead of localhost necessary for Docker
}

func launchBench(c *gin.Context) {
	stanlog.Info("received request for inserting " + c.Param("recordscount") + " records on " + c.Param("policy") + " storage")
	policy := c.Param("policy")
	count, err := strconv.Atoi(c.Param("recordscount"))
	if err != nil {
		c.JSON(400, gin.H{"Error": "Incorrect format of records count"})
		return
	}
	c.String(200, "Insertion request received. Please consult log for results.")
	insertRecords(count, policy)
}

func launchBatch(c *gin.Context) {
	stanlog.Info("received request for " + c.Param("iterations") + " iterations of inserting " + c.Param("recordscount") + " records on " + c.Param("policy") + " storage")
	policy := c.Param("policy")
	count, err := strconv.Atoi(c.Param("recordscount"))
	if err != nil {
		c.JSON(400, gin.H{"Error": "Incorrect format of records count"})
		return
	}
	iterations, err := strconv.Atoi(c.Param("iterations"))
	if err != nil {
		c.JSON(400, gin.H{"Error": "Incorrect format of records count"})
		return
	}
	c.String(200, "Batch insertion request received. Please consult log for results.")
	results = nil
	for i := 0; i < iterations; i++ {
		insertRecords(count, policy)
	}
	stanlog.Info("Results contain " + strconv.Itoa(len(results)) + " elements of " + c.Param("recordscount") + " records written to " + c.Param("policy"))
	stanlog.Info("Values : " + fmt.Sprint(results))
}

func connect() (driver.Conn, error) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{config.Clickhouse.Server + ":" + config.Clickhouse.Port},
			ClientInfo: clickhouse.ClientInfo{
				Products: []struct {
					Name    string
					Version string
				}{
					{Name: "Perf Test", Version: "0.1"},
				},
			},

			Debugf: func(format string, v ...interface{}) {
				fmt.Printf(format, v)
			},
		})
	)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return conn, nil
}

func insertRecords(nbrecords int, storagePolicy string) {
	conn, err := connect()
	if err != nil {
		stanlog.Critical(err.Error())
	}

	tableName := "perftest"
	ctx := context.Background()
	conn.Exec(ctx, `DROP TABLE IF EXISTS `+tableName)
	err = conn.Exec(context.Background(), `
	CREATE table `+tableName+`
	(
		"@timestamp" Int32 EPHEMERAL 0,
		data Nested(machine String, user String),
		timestamp    DateTime DEFAULT toDateTime("@timestamp")
	) ENGINE = MergeTree() ORDER BY (timestamp)	SETTINGS storage_policy = '`+storagePolicy+`'
	`)
	if err != nil {
		stanlog.Error(err.Error())
		return
	}
	stanlog.Info("Table " + tableName + " created with storagePolicy " + storagePolicy)

	conn.Exec(ctx, `SET input_format_import_nested_json = 1;`)

	stanlog.Info("Inserting " + strconv.Itoa(nbrecords) + " records")
	start := time.Now()
	for i := 0; i < nbrecords; i++ {
		err = conn.Exec(ctx, `
		INSERT INTO `+tableName+` ("@timestamp", data.machine, data.user)
		FORMAT JSONEachRow
			{"@timestamp":897819077, "data":{"machine":["wks `+strconv.Itoa(i)+`"], "user":["usr `+strconv.Itoa(i)+`"] }}
		`)
		if err != nil {
			stanlog.Error("Error in iteration " + strconv.Itoa(i) + err.Error())
		}
	}
	stop := time.Now()
	diff := stop.Sub(start)
	results = append(results, diff.Milliseconds())

	stanlog.Info("Inserted " + strconv.Itoa(nbrecords) + " records on " + storagePolicy + " storage in " + strconv.FormatInt(diff.Milliseconds(), 10) + " milliseconds")
}
