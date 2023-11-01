// Copyright 2023 daz-3ux(Daz) <daz-3ux@proton.me>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/Daz-3ux/dCache.

package register

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
)

func TestRegister(t *testing.T) {
	cli, _ := clientv3.New(DefaultEtcdConfig)

	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = etcdAdd(cli, resp.ID, "test", "localhost:6324")
	if err != nil {
		t.Fatalf(err.Error())
	}
}
