import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import * as serviceWorker from './serviceWorker';
import ChalkClient from './chalk/ChalkClient';
import { getActiveSession } from './services/sessions';
import { Bootstrap } from './bootstrap';

// Allow click-outside to work on iOS.
if ('ontouchstart' in document.documentElement) {
  document.body.style.cursor = 'pointer';
}

// Load bootstrap (embedded in page html)
function getBootstrap(): Bootstrap {
  interface Window {
    bootstrap: Bootstrap;
  }

  if (!(window instanceof Window)) {
    throw 'No bootstrap present in window.';
  }

  const win = window as unknown as Window;
  return win.bootstrap;
}

const bootstrap = getBootstrap();
const api_url = `//${bootstrap.api_host}`;

const chalk = new ChalkClient(api_url);

ReactDOM.render(
  <App
    currentPageId={bootstrap.page_id}
    checkConnection={() => chalk.checkConnection()}
    createVariable={(pageId, name, formula) => chalk.createVariable(pageId, name, formula)}
    updateVariable={(id, formula) => chalk.updateVariable(id, formula)}
    renameVariable={(id, name) => chalk.renameVariable(id, name)}
    getPageVariables={(pageId) => chalk.getPageVariables(pageId)}
    getSession={() => getActiveSession(chalk)} />,
  document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: http://bit.ly/CRA-PWA
serviceWorker.unregister();
