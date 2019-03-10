import axios from 'axios';
import {List} from 'immutable';

import {VariableState} from '../domain';
import {ApiResult, parseApiResultObject} from '../resolver';
import { Result } from '../domain/resolver';


export async function createVariable(apiUrl: string, pageId: string, name: string, formula: string): Promise<VariableState> {
  const {data} = await axios.post(apiUrl+'/variables', {
    page: pageId,
    name,
    formula,
  });

  const payload: VariableMutationResponse = data;

  if (payload.error !== undefined) {
    throw payload.error;
  }

  if (payload.state === undefined) {
    throw 'Expected variable state';
  }

  return parseVariableStateResponse(payload.state);
}

export async function getPageVariables(apiUrl: string, pageId: string): Promise<List<VariableState>> {
  const {data} = await axios.get(`${apiUrl}/pages/${pageId}/variables`);
  const payload: GetPageVariablesResponse = data;

  return List(payload.variables).map(parseVariableStateResponse);
}

export async function renameVariable(apiUrl: string, id: string, name: string): Promise<VariableState> {
  const {data} = await axios.post(apiUrl+`/variables/${id}`, {
    name,
  });

  const payload: VariableMutationResponse = data;

  if (payload.error !== undefined) {
    throw payload.error;
  }

  if (payload.state === undefined) {
    throw 'Expected variable state';
  }

  return parseVariableStateResponse(payload.state);
}

export async function updateVariable(apiUrl: string, id: string, formula: string): Promise<VariableState> {
  const {data} = await axios.post(apiUrl+`/variables/${id}`, {
    formula,
  });

  const payload: VariableMutationResponse = data;

  if (payload.error !== undefined) {
    throw payload.error;
  }

  if (payload.state === undefined) {
    throw 'Expected variable state';
  }

  return parseVariableStateResponse(payload.state);
}

function parseVariableStateResponse(state: VariableStateResponse): VariableState {
  let result: Result;

  if (state.result.error !== undefined) {
    result = {
      resultType: 'error',
      message: state.result.error,
    };
  } else {
    if (state.result.result === undefined) {
      throw 'Expected variable result';
    }
    result = parseApiResultObject(state.result.result);
  }

  return {
    id: state.id,
    name: state.name,
    formula: state.formula,
    result,
  }
}

interface GetPageVariablesResponse {
  variables: ReadonlyArray<VariableStateResponse>,
}

interface VariableMutationResponse {
  error?: string,
  state?: VariableStateResponse,
}

interface VariableStateResponse {
  id: string,
  name: string,
  formula: string,
  result: ApiResult,
}

interface GetVariablesResponse {
  variables: ReadonlyArray<VariableStateResponse>,
}
