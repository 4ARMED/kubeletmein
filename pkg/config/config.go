// Copyright © 2018 Marc Wickenden <marc@4armed.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package config

// Config holds configuration values for the various operations
type Config struct {
	Provider        string
	SkipBootstrap   bool
	BootstrapConfig string
	CaCertPath      string
	KubeletKeyPath  string
	KubeletCertPath string
	KubeletToken    string
	KubeConfig      string
	CertDir         string
	NodeName        string
	KubeAPIServer   string
	MetadataFile    string
	Region          string
}
