//
// Copyright © 2016 Ikey Doherty <ikey@solus-project.com>
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

// UserInfo is required for ypkg builds, to set the .solus/package internally
// and propagate the author details.
type UserInfo struct {
	Name  string
	Email string
}

// GetUserInfo will always succeed, as it will use a fallback policy until it
// finally comes up with a valid combination of name/email to use.
func GetUserInfo() *UserInfo {
	return nil
}
