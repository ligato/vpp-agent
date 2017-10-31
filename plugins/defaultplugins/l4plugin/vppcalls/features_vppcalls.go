// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/bin_api/session"
)

// EnableL4Features sets L4 feature flag on VPP to true
func EnableL4Features(log logging.Logger, vppChan *govppapi.Channel) error {
	log.Debug("Enabling L4 features")

	req := &session.SessionEnableDisable{
		IsEnable: 1,
	}
	reply := &session.SessionEnableDisableReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		log.WithFields(logging.Fields{"Error": err, "L4Features": true}).Error("Error while enabling L4 features")
		return err
	}
	if reply.Retval != 0 {
		log.WithField("Return value", reply.Retval).Error("Unexpected return value")
		return err
	}
	log.Debug("L4 features enabled.")

	return nil
}

// DisableL4Features sets L4 feature flag on VPP to false
func DisableL4Features(log logging.Logger, vppChan *govppapi.Channel) error {
	log.Debug("Disabling L4 features")

	req := &session.SessionEnableDisable{
		IsEnable: 0,
	}
	reply := &session.SessionEnableDisableReply{}
	err := vppChan.SendRequest(req).ReceiveReply(reply)
	if err != nil {
		log.WithFields(logging.Fields{"Error": err, "L4Features": true}).Error("Error while disabling L4 features")
		return err
	}
	if reply.Retval != 0 {
		log.WithField("Return value", reply.Retval).Error("Unexpected return value")
		return err
	}
	log.Debug("L4 features disabled.")

	return nil
}
