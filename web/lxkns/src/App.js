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

import React, { useEffect, useState } from 'react';
import './App.css';
import Namespace from './Namespace';
import { postDiscovery, namespaceIdOrder } from './model';

function App() {
  const [allns, setallns] = useState({namespaces: {}, processes: {}});

  useEffect(() => {
    namespaceDiscovery();
  }, []);
  
  const namespaceDiscovery = async () => {
    const response = await fetch(
      'http://' + window.location.hostname + ':5010/api/namespaces');
    const jsondata = await response.json();
    setallns(postDiscovery(jsondata));
  };

  const nslist = Object.values(allns.namespaces)
    .filter(ns => ns.type === "user")
    .sort(namespaceIdOrder)
    .map(ns => 
      <li key={ns.nsid.toString()}><Namespace ns={ns}/></li>
    );

  return (
    <div className="App">
      <header className="App-header">
        <ul>
          {nslist}
        </ul>
      </header>
    </div>
  );
}

export default App;
