import axios from 'axios';
import {List} from 'immutable';

import {createSession, getSession} from './sessions';
import {createVariable, renameVariable, updateVariable, getPageVariables} from './variables';
import {SessionState, VariableState} from './domain';

class ChalkClient {
  apiUrl: string;

  constructor(apiUrl: string) {
    this.apiUrl = apiUrl;
  }

  createSession(): Promise<SessionState> {
    return createSession(this.apiUrl);
  }

  getSession(sessionId: string): Promise<SessionState | null> {
    return getSession(this.apiUrl, sessionId);
  }

  createVariable(pageId: string, varName: string, formula: string): Promise<VariableState> {
    return createVariable(this.apiUrl, pageId, varName, formula);
  }

  renameVariable(id: string, varName: string): Promise<VariableState> {
    return renameVariable(this.apiUrl, id, varName);
  }

  updateVariable(id: string, formula: string): Promise<VariableState> {
    return updateVariable(this.apiUrl, id, formula);
  }

  getPageVariables(pageId: string): Promise<List<VariableState>> {
    return getPageVariables(this.apiUrl, pageId);
  }

  checkConnection(): Promise<{}> {
    return axios.get(this.apiUrl + '/health');
  }
}

export default ChalkClient;
