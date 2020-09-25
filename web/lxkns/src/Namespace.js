// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

import React from 'react';
import classNames from 'classnames';
import DirectionsRunIcon from '@material-ui/icons/DirectionsRun';
import LinkIcon from '@material-ui/icons/Link';
import PersonIcon from '@material-ui/icons/Person';
import StorageIcon from '@material-ui/icons/Storage';
import SettingsEthernetIcon from '@material-ui/icons/SettingsEthernet';
import PhoneInTalkIcon from '@material-ui/icons/PhoneInTalk';
import SpeedIcon from '@material-ui/icons/Speed';
import TimerIcon from '@material-ui/icons/Timer';
import DnsIcon from '@material-ui/icons/Dns';

const icons = {
    "cgroup": <SpeedIcon fontSize="inherit"/>,
    "ipc": <PhoneInTalkIcon fontSize="inherit"/>,
    "mnt": <StorageIcon fontSize="inherit"/>,
    "net": <SettingsEthernetIcon fontSize="inherit"/>,
    "pid": <DirectionsRunIcon fontSize="inherit"/>,
    "user": <PersonIcon fontSize="inherit"/>,
    "uts": <DnsIcon fontSize="inherit"/>,
    "time": <TimerIcon fontSize="inherit"/>
};

// Component Namespace renders information about a particular namespace: type
// and ID, as well as the most senior process with its name, or a bind-mounted
// reference. This component never renders any child namespaces (of PID and
// user namespaces).
const Namespace = (props) => {
    const process = (props.ns.ealdorman &&
        <span className="processinfo"><DirectionsRunIcon fontSize="inherit"/>
            process <span className="processname">"{props.ns.ealdorman.name}"</span>
        ({props.ns.ealdorman.pid})
      </span>) || (props.ns.reference &&
            <span className="bindmount"><LinkIcon fontSize="inherit"/>
                bind-mounted at <span className="bindmount">"{props.ns.reference}"</span>
      </span>) || "";

    const cgroup = (props.ns.cgroup &&
        <span className="cgroupinfo">
            controlled by "{props.ns.cgroup}"
      </span>) || "";

    return <span className={classNames("namespace", props.ns.type)}>
        {icons[props.ns.type]}&nbsp;
        <span className="pill">{props.ns.type}:[{props.ns.nsid}]</span>
        {process}
        {cgroup}
    </span>;
};

export default Namespace;
