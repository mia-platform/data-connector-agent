// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mapper

/*
{
  "id": "id1",
	"fields": {
		"foo": "bar",
		"baz": "qux"
	}
}

to

{
	"identifier": "id1",
	"foo": "bar",
	"pippo": "qux"
}
*/

type IMapper[T any] interface {
	Map(data map[string]any) (map[string]any, error)
}

type Mapper[T any] struct {
}

// func (m *Mapper[T]) Map(data []byte) (map[string]any, error) {
// 	// TODO: implement the mapping logic
// 	return data, nil
// }

// func NewMapper(_ map[string]any) IMapper {
// 	return &Mapper{}
// }