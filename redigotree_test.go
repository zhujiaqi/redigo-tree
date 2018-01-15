package redigotree

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRedisTree(t *testing.T) {
	TInsert("test", "1", "2", nil)
	TInsert("test", "1", "4", nil)
	TInsert("test", "1", "3", map[string]string{"before": "4"})
	TInsert("test", "2", "5", map[string]string{"index": "1000"})
	TInsert("test", "5", "4", nil)

	//fmt.Println(redis.TChildren("test", "1", nil))

	parents1 := TParents("test", "5") //["2"]
	parents2 := TParents("test", "1") // []
	parents3 := TParents("test", "4") //["5", "1"]

	Convey("TestRedisTree", t, func() {
		Convey("Test parents1", func() {
			//fmt.Println(parents1)
			So(len(parents1), ShouldEqual, 1)
		})
		Convey("Test parents2", func() {
			//fmt.Println(parents2)
			So(len(parents2), ShouldEqual, 0)
		})
		Convey("Test parents3", func() {
			//fmt.Println(parents3)
			So(len(parents3), ShouldEqual, 2)
		})
	})
	TDestroy("test", "1")
}

func TestMoveRedisTree(t *testing.T) {
	TInsert("test", "root", "1", nil)
	TInsert("test", "1", "2", nil)
	TInsert("test", "2", "3", nil)
	TInsert("test", "3", "4", nil)
	TInsert("test", "4", "5", nil)

	fmt.Println("TestMoveRedisTree:1", TPath("test", "root", "5"))

	TMoveChildren("test", "1", "root", "APPEND")
	TMoveChildren("test", "2", "root", "APPEND")
	TMoveChildren("test", "3", "root", "APPEND")

	fmt.Println("TestMoveRedisTree:2", TPath("test", "root", "5"))

	fmt.Println("TestMoveRedisTree:3", TChildren("test", "root", nil))

	TDestroy("test", "root")
}

func TestTRem(t *testing.T) {
	TInsert("test", "1", "2", nil)
	TInsert("test", "1", "4", nil)
	TInsert("test", "1", "3", map[string]string{"before": "4"})
	TInsert("test", "2", "5", map[string]string{"index": "1000"})
	TInsert("test", "5", "4", nil)

	fmt.Println("TestTRem:1", TChildren("test", "1", nil))

	fmt.Println("TestTRem:2", TRem("test", "5", 0, "4"))

	fmt.Println("TestTRem:3", TChildren("test", "1", nil))

	TDestroy("test", "1")
}

func TestTPrune(t *testing.T) {
	TInsert("test", "1", "5", nil)
	TInsert("test", "1", "4", nil)
	TInsert("test", "6", "5", nil)

	fmt.Println("TestTPrune:1", TChildren("test", "1", nil))

	fmt.Println("TestTPrune:2", TChildren("test", "6", nil))

	fmt.Println("TestTPrune:3", TPrune("prune", "1"))

	fmt.Println("TestTPrune:4", TChildren("test", "1", nil))

	fmt.Println("TestTPrune:5", TChildren("test", "6", nil))
	TDestroy("test", "1")
}
