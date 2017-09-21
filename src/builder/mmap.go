//
// Copyright © 2017 Solus Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package builder

import (
	"os"
	"syscall"
)

// An MmapFile is used to easily wrap the syscall mmap() functions to be readily
// usable from golang.
// This helps heaps (pun intended) when it comes to computing the hash sum for
// very large files, such as the index and eopkg files, in a zero copy fashion.
type MmapFile struct {
	f    *os.File
	Data []byte
	len  int64
	m    bool
}

// MapFile will attempt to mmap() the input file
func MapFile(path string) (*MmapFile, error) {
	var err error
	ret := &MmapFile{}
	ret.f, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	st, err := ret.f.Stat()
	if err != nil {
		ret.f.Close()
		return nil, err
	}
	ret.len = st.Size()
	ret.Data, err = syscall.Mmap(int(ret.f.Fd()), 0, int(ret.len), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		ret.f.Close()
		return nil, err
	}
	ret.m = true
	return ret, nil
}

// Close will close the previously mmapped filed
func (m *MmapFile) Close() error {
	var err error
	if m.f == nil {
		return nil
	}
	if m.m {
		err = syscall.Munmap(m.Data)
		m.m = false
	}
	m.f.Close()
	m.f = nil
	return err
}
