package redigotree

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
)

type TreeNode struct {
	Node     string     `json:"node"`
	HasChild bool       `json:"hasChild"`
	Children []TreeNode `json:"children"`
}

var treeCommands map[string]string

var head = strings.Join(Filter(strings.Split(loadScript("_head"), "\n"), isNotCommend), " ") + " "
var deleteReference = strings.Join(Filter(strings.Split(loadScript("_delete_reference"), "\n"), isNotCommend), " ") + " "
var getPath = strings.Join(Filter(strings.Split(loadScript("_get_path"), "\n"), isNotCommend), " ") + " "

func loadScript(name string) string {
	_, file, _, _ := runtime.Caller(1)
	path := filepath.Join(filepath.Dir(file), "lua")
	b, err := ioutil.ReadFile(path + "/" + name + ".lua")
	if err != nil {
		panic(fmt.Sprintf("Load Lua Script Failed: %s", err))
	}
	str := string(b)
	str = strings.Join(Filter(strings.Split(str, "\n"), isNotCommend), " ") + " "
	return str
}

func isNotCommend(line string) bool {
	if line == "" {
		return false
	}
	return strings.TrimSpace(line)[0:2] != "--"
}

func getLuaScript(command string) string {
	lua := loadScript(command)
	if command == "trem" || command == "tdestroy" || command == "tmrem" || command == "tprune" {
		lua = deleteReference + lua
	} else if command == "tpath" || command == "tinsert" || command == "tmovechildren" {
		lua = getPath + lua
	}
	return head + lua
}

var commands = []string{
	"tinsert", "tchildren", "tparents", "tpath", "trem", "tmrem",
	"tdestroy", "texists", "trename", "tprune", "tmovechildren",
}

func init() {
	treeCommands = func(vs []string, f func(string) string) map[string]string {
		vsm := make(map[string]string, len(vs))
		for _, v := range vs {
			vsm[v] = f(v)
		}
		return vsm
	}(commands, getLuaScript)
}

func convertNode(t interface{}) TreeNode {
	var ret TreeNode
	items, ok := t.([]interface{})
	if !ok {
		return TreeNode{}
	}
	ret = TreeNode{
		B2S(items[0].([]uint8)),
		items[1].(int64) > 0,
		nil,
	}
	if len(items) > 2 {
		ret.Children = []TreeNode{}
		for i := 2; i < len(items); i++ {
			child := convertNode(items[i])
			if child.Node != "" {
				ret.Children = append(ret.Children, child)
			}
		}
	}
	return ret
}

func callTreeApi(command string, args []interface{}) (reply interface{}, err error) {
	if lua, ok := treeCommands[command]; ok {
		redisClient := redisClient.Get()
		defer redisClient.Close()
		var getScript = redis.NewScript(1, lua)
		return getScript.Do(redisClient, args...)
	}
	// log.Error("RedisTree Command Not Exist")
	return nil, errors.New("RedisTree Command Not Exist")
}

func TInsert(key string, parent string, node string, options map[string]string) int {
	var reply interface{}
	var err error
	args := []interface{}{key, parent, node}
	if option, ok := options["index"]; ok {
		args = append(args, "INDEX", option) //Index, Returned by Insert
	} else if option, ok := options["before"]; ok {
		args = append(args, "BEFORE", option) //Insert Before Node
	} else if option, ok := options["after"]; ok {
		args = append(args, "AFTER", option) //Insert After Node
	} else {
		args = append(args, "INDEX", -1) //Default: Insert At Last Pos
	}
	reply, err = callTreeApi("tinsert", args)
	if err != nil {
		log.Error("RedisTree TInsert Error: ", err)
		return -1
	}
	/* log.WithFields(log.Fields{
		"key": key,
		"parent": parent,
		"node": node,
		"options": options,
	}).Debug("RedisTree TInsert")
	*/
	insertIndex, _ := redis.Int(reply, err)
	return insertIndex
}

func TChildren(key string, node string, options map[string]string) []TreeNode {
	var reply interface{}
	var err error
	args := []interface{}{key, node}
	if option, ok := options["level"]; ok {
		args = append(args, "LEVEL", option)
	}
	reply, err = callTreeApi("tchildren", args)
	if err != nil {
		log.Error("RedisTree TChildren Error: ", err)
		return nil
	}
	/*log.WithFields(log.Fields{
		"key": key,
		"node": node,
		"options": options,
	}).Debug("RedisTree TChildren")
	*/
	items, err := redis.Values(reply, err)
	return func(vs []interface{}, f func(interface{}) TreeNode) []TreeNode {
		vsm := make([]TreeNode, len(vs))
		for i, v := range vs {
			result := f(v)
			if result.Node != "" {
				vsm[i] = f(v)
			}
		}
		// log.Debug("RedisTree TChildren: returned")
		return vsm
	}(items, convertNode)
}

func TParents(key string, node string) []string {
	var reply interface{}
	var err error
	args := []interface{}{key, node}
	reply, err = callTreeApi("tparents", args)
	if err != nil {
		log.Error("RedisTree TParents Error: ", err)
		return nil
	}
	/*log.WithFields(log.Fields{
		"key": key,
		"node": node,
	}).Debug("RedisTree TParents")
	*/
	parents, _ := redis.Strings(reply, err)
	return parents
}

func TPath(key string, node string, newNode string) []string {
	var reply interface{}
	var err error
	args := []interface{}{key, node, newNode}
	reply, err = callTreeApi("tpath", args)
	if err != nil {
		log.Error("RedisTree TPath Error: ", err)
		return []string{}
	}
	/*log.WithFields(log.Fields{
		"key": key,
		"node": node,
		"newnode": newNode,
	}).Debug("RedisTree TPath")
	*/
	path, _ := redis.Strings(reply, err)
	return path
}

func TRem(key string, parent string, count int, node string) int {
	var reply interface{}
	var err error
	args := []interface{}{key, parent, count, node}
	reply, err = callTreeApi("trem", args)
	if err != nil {
		log.Error("RedisTree TRem Error: ", err)
		return -1
	}
	num, _ := redis.Int(reply, err)
	// TREM returns count of remaining nodes in the parent.
	return num
}

func TMrem(key string, node string, options map[string]string) interface{} {
	var reply interface{}
	var err error
	args := []interface{}{key, node}
	if option, ok := options["not"]; ok {
		args = append(args, "NOT", option) //Exclude a parent
	}
	reply, err = callTreeApi("tmrem", args)
	if err != nil {
		log.Error("RedisTree TMrem Error: ", err)
		return nil
	}
	/* log.WithFields(log.Fields{
		"key": key,
		"node": node,
		"options": options,
	}).Debug("RedisTree TMrem")
	*/
	return reply
}

func TDestroy(key string, node string) int {
	var reply interface{}
	var err error
	args := []interface{}{key, node}
	reply, err = callTreeApi("tdestroy", args)
	if err != nil {
		log.Error("RedisTree TDestroy Error: ", err)
		return -1
	}
	/*
		log.WithFields(log.Fields{
			"key":  key,
			"node": node,
		}).Debug("RedisTree TDestroy")
	*/
	numberOfNodesDestroyed, _ := redis.Int(reply, err)
	return numberOfNodesDestroyed
}

func TExists(key string, node string) bool {
	var reply interface{}
	var err error
	args := []interface{}{key, node}
	reply, err = callTreeApi("texists", args)
	if err != nil {
		log.Error("RedisTree TExists Error: ", err)
		return false
	}
	/*log.WithFields(log.Fields{
		"key": key,
		"node": node,
	}).Debug("RedisTree TExists")*/
	exists, _ := redis.Bool(reply, err)
	return exists
}

func TRename(key string, node string, newNode string) bool {
	var reply interface{}
	var err error
	args := []interface{}{key, node, newNode}
	reply, err = callTreeApi("trename", args)
	if err != nil {
		log.Error("RedisTree TRename Error: ", err)
		return false
	}
	/*log.WithFields(log.Fields{
		"key": key,
		"node": node,
		"newnode": newNode,
	}).Debug("RedisTree TRename")*/
	exists, _ := redis.Bool(reply, err)
	return exists
}

func TPrune(key string, node string) bool {
	var reply interface{}
	var err error
	args := []interface{}{key, node}
	reply, err = callTreeApi("tprune", args)
	if err != nil {
		log.Error("RedisTree TPrune Error: ", err)
		return false
	}
	pruned, _ := redis.String(reply, err)
	return pruned == "OK"
}

func TMoveChildren(key string, source string, target string, op string) int {
	if op != "PREPEND" && op != "APPEND" {
		op = "APPEND"
	}
	var reply interface{}
	var err error
	args := []interface{}{key, source, target, op}
	reply, err = callTreeApi("tmovechildren", args)
	if err != nil {
		log.Error("RedisTree TMoveChildren Error: ", err)
		return -1
	}
	/*
		log.WithFields(log.Fields{
			"key":    key,
			"source": source,
			"target": target,
			"op":     op,
		}).Debug("RedisTree TMoveChildren")
	*/
	sourceListSize, _ := redis.Int(reply, err)
	return sourceListSize
}
