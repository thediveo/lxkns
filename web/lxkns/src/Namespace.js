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
import { namespaceNameTypeIdOrder } from './model';

// Namespace renders an individual namespace component. If the namespace is a
// user namespace, then it additionally renders all owned namespaces as child
// elements.
const Namespace = (props) => {
    const pillcls = classNames("pill", props.ns.type);
    const process = (props.ns.ealdorman &&
        <span className="processinfo">
            process "{props.ns.ealdorman.name}"
        ({props.ns.ealdorman.pid})
      </span>) || (props.ns.reference &&
            <span className="bindmount">
                bind-mounted at "{props.ns.reference}"
      </span>) || "";

    const cgroup = (props.ns.cgroup &&
        <span className="cgroupinfo">
            controlled by "{props.ns.cgroup}"
      </span>) || "";

    const tens = (props.ns.tenants !== undefined &&
        props.ns.tenants.sort(namespaceNameTypeIdOrder)
            .map(ns =>
                <li key={ns.nsid.toString()}><Namespace ns={ns} /></li>
            )) || "";
    const tenants = (tens && <ul>{tens}</ul>) || ""

    return <span className="namespace">
        <span className={pillcls}>{props.ns.type}:[{props.ns.nsid}]</span>
        {process}
        {cgroup}
        {tenants}
    </span>;
}

export default Namespace;
