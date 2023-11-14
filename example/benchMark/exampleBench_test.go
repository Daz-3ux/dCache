// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package benchMark

import (
	"fmt"
	"github.com/Daz-3ux/dazCache/dCache"
	"testing"
)

func BenchmarkGetDazScore(b *testing.B) {
	// 模拟MySQL数据库
	mysql := map[string]string{
		"daz":     "666",
		"realdaz": "777",
		"fakedaz": "888",
	}
	group := dCache.NewGroup("scores", 2<<10, dCache.GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	_, err := group.Get("daz")
	if err != nil {
		b.Fatalf("Error getting value: %s", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := group.Get("daz")
		if err != nil {
			b.Fatalf("Error getting value: %s", err)
		}
	}
}
