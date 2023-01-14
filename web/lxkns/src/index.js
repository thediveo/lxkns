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

import flat from 'core-js/features/array/flat'
import React from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './app'

// Import only the necessary Roboto fonts, so they are available "offline"
// without CDN.
import '@fontsource/roboto/300.css'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import '@fontsource/roboto-mono/400.css'

// HACK: for reasons yet unknown to mankind, the usual direct import of
// 'core-js/features/array/flat' doesn't correctly fix missing Array.flat() on
// some browsers; however, a non-polluting import with explicit pollution then
// works. 
if (Array.flat === undefined) {
	Array.flat = flat
}

// Allow development version to temporarily drop strict mode in order to see
// performance without strict-mode double rendering.
const container = document.getElementById('root');
createRoot(container).render(
	process.env.REACT_APP_UNSTRICT
		? <App />
		: <React.StrictMode><App /></React.StrictMode>
);
