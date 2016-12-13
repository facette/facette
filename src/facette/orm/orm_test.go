package orm

import (
	"database/sql"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Foo struct {
	ID          uint      `orm:"primary_key;auto_increment"`
	Name        string    `orm:"type:varchar(64);not_null;unique"`
	Description string    `orm:"column:desc"`
	Time        time.Time `orm:"not_null;default:now();index"`
	Ignored     string    `orm:"-"`
	Foo         *Foo      `orm:"foreign_key:ID"`
	BarA        Bar       `orm:"foreign_key:ID"`
	BarB        *Bar      `orm:"foreign_key:ID"`
	Bar         []Bar     `orm:"foreign_key:ID"`
	Baz         []*Baz    `orm:"foreign_key:ID"`
	FooTypes
}

type FooTypes struct {
	TypeBool    bool
	TypeInt     int
	TypeInt8    int8
	TypeInt16   int16
	TypeInt32   int32
	TypeInt64   int64 `orm:"index:value64"`
	TypeUint    uint
	TypeUint8   uint8
	TypeUint16  uint16
	TypeUint32  uint32
	TypeUint64  uint64  `orm:"index:value64"`
	TypeFloat32 float32 `orm:"index:float"`
	TypeFloat64 float64 `orm:"index:float,value64"`
}

type Bar struct {
	ID    int `orm:"primary_key"`
	Value string
}

type Baz struct {
	ID    int `orm:"primary_key"`
	Value string
}

func testORM(driver, dsn string, t *testing.T, useTx bool) {
	// Open database
	db, err := Open(driver, dsn)
	if err != nil {
		t.Fatalf("failed to open database: %s", err)
	}
	defer db.Close()

	db.SetLogger(log.New(os.Stdout, "", log.LstdFlags))

	if useTx {
		db = db.Begin()
		defer db.Commit()
	}

	if err := db.Migrate(Foo{}).Error(); err != nil {
		t.Log(err)
		t.Fail()
	}

	now := time.Now().Round(time.Second).UTC()

	parent := Foo{
		Name: "parent",
		Time: now.AddDate(0, 0, -1),
	}

	foo := Foo{
		Name:        "foo",
		Description: "A great description here",
		Time:        now,
		Ignored:     "This text should be ignored",
		Foo:         &parent,
		BarA: Bar{
			ID:    123,
			Value: "value1",
		},
		BarB: &Bar{
			ID:    456,
			Value: "value2",
		},
		Bar: []Bar{
			Bar{101, "value3"},
			Bar{102, "value4"},
			Bar{103, "value5"},
		},
		Baz: []*Baz{
			&Baz{104, "value6"},
			&Baz{105, "value7"},
		},
		FooTypes: FooTypes{
			TypeBool:    true,
			TypeInt:     1,
			TypeInt8:    1,
			TypeInt16:   1,
			TypeInt32:   1,
			TypeInt64:   1,
			TypeUint:    1,
			TypeUint8:   1,
			TypeUint16:  1,
			TypeUint32:  1,
			TypeUint64:  1,
			TypeFloat32: 1.23,
			TypeFloat64: 45.67,
		},
	}

	if err := db.Save(&foo).Error(); err != nil {
		t.Log(err)
		t.Fail()
	}

	foo.Description += " (updated)"
	foo.Bar = foo.Bar[1:]
	foo.FooTypes.TypeBool = false

	if err := db.Save(&foo).Error(); err != nil {
		t.Log(err)
		t.Fail()
	}

	foo.Ignored = ""

	fetchOne := Foo{}
	if err := db.Where("id = ?", 2).Find(&fetchOne).Error(); err != nil {
		t.Log(err)
		t.Fail()
	} else if !reflect.DeepEqual(foo, fetchOne) {
		t.Logf("\nExpected %#v\nbut got  %#v\n", foo, fetchOne)
		t.Fail()
	}

	fetchMany := []Foo{}
	fetchManyExpexted := []Foo{parent, foo}

	if err := db.OrderBy("id").Find(&fetchMany).Error(); err != nil {
		t.Log(err)
	} else if !reflect.DeepEqual(fetchManyExpexted, fetchMany) {
		t.Logf("\nExpected %#v\nbut got  %#v\n", fetchManyExpexted, fetchMany)
		t.Fail()
	}

	fetchMany = []Foo{}
	fetchManyExpexted = []Foo{parent}
	if err := db.Where("id IN (?)", []uint{1, 3}).OrderBy("id").Find(&fetchMany).Error(); err != nil {
		t.Log(err)
	} else if !reflect.DeepEqual(fetchManyExpexted, fetchMany) {
		t.Logf("\nExpected %#v\nbut got  %#v\n", fetchManyExpexted, fetchMany)
		t.Fail()
	}

	if err := db.Delete(&foo).Error(); err != nil {
		t.Log(err)
		t.Fail()
	}

	if err := db.Find(&fetchOne).Error(); err != nil && err != sql.ErrNoRows {
		t.Log(err)
		t.Fail()
	}

	if err := db.DropTable("foos_bars_assoc", "foos_bazs_assoc", Foo{}, Bar{}, Baz{}).Error(); err != nil {
		t.Log(err)
		t.Fail()
	}
}
