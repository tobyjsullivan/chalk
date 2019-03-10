import {List} from 'immutable';

import {Result} from './resolver';

export interface PageState {
  id: string,
}

export interface SessionState {
  id: string,
  pages: List<PageState>,
};

export interface VariableState {
  id: string,
  name: string,
  formula: string,
  result: Result,
}
