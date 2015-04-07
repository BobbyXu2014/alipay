package conver

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
)

var c Convertor

type Convertor struct{}

// Do 自定义解析器
// 将json.Unmarshal过的map值映射到结构体
func Do(o interface{}, params map[string]interface{}) error {
	return c.doconv(o, params)
}

// valid 判断参数合法性
func (c *Convertor) doconv(o interface{}, params map[string]interface{}) error {
	// log.Printf("type : %s,value: %s", reflect.TypeOf(o).Kind(), params)
	rv := reflect.ValueOf(o)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New(fmt.Sprintf("%v can not be assign ", o))
	}
	return c.inject(rv, params)
}

// inject 注入值
func (c *Convertor) inject(v reflect.Value, params map[string]interface{}) (err error) {

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	// log.Println(v.IsValid(), v.Kind(), params)
	if !v.IsValid() {
		return errors.New(fmt.Sprintf("%s is invalid", v))
	}

	// 拿到与数据匹配的字段名
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	nv := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("align")
		for k, v := range params {
			if tag == k {
				nv[f.Name] = v
			}
		}
	}
	if len(nv) == 0 {
		// 第一轮没有匹配的，拿到value后进行第二轮
		for _, v := range params {
			if obj, ok := v.(map[string]interface{}); ok {
				for ok, ov := range obj {
					nv[ok] = ov
				}
			}
		}
		if len(nv) != 0 {
			return c.inject(v, nv)
		}
	}
	// log.Println(nv)
	for k, value := range nv {

		f := reflect.Value{}
		//TODO 待扩展类型
		if v.Kind() == reflect.Ptr {
			f = v.Elem().FieldByName(k)
			// log.Println(f.Kind(), k)
		} else {
			f = v.FieldByName(k)
		}

		switch f.Kind() {
		case reflect.String:
			if str, ok := value.(string); ok {
				// log.Println(f.Kind(), value)
				f.SetString(str)
			}
			// if value is float64,just convet it to string
			if float, ok := value.(float64); ok {
				// or fmt.Sprintf("%.0f", float)
				f.SetString(strconv.FormatFloat(float, 'f', -1, 64))
			}
		case reflect.Struct:
			// 如果是struct，直接进行递归
			if m, ok := value.(map[string]interface{}); ok {
				// f.Set(reflect.New(f.Type()))
				c.inject(f, m)
			}
		case reflect.Ptr:
			// 如果是ptr，判断是否非法，是，进行初始化
			if m, ok := value.(map[string]interface{}); ok {
				if !f.Elem().IsValid() {
					// 用new就变成另一个地址
					// f = reflect.New(f.Type().Elem())
					f.Set(reflect.New(f.Type().Elem()))
					// log.Printf("%s ,,tes", f)
				}
				c.inject(f, m)
			}
		}
	}
	return nil
}
