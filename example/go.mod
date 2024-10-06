module github.com/aneshas/eventstore-example

go 1.23

require (
	github.com/aneshas/eventstore v0.3.0
	gorm.io/driver/sqlite v1.5.0
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	gorm.io/gorm v1.25.0 // indirect
)

replace github.com/aneshas/eventstore => ../../eventstore
