import axios from 'axios';

import {Result, None} from '../domain/resolver';

export function executeFormula(apiUrl: string, formula: string): Promise<Result> {
  return axios.post(apiUrl+'/execute', {
    formula,
  }).then((resp) => {
    const payload: ApiResult = resp.data;

    if (payload.error) {
      throw payload.error;
    }

    if (!payload.result) {
      // Empty result
      return None;
    }
    return parseApiResultObject(payload.result);
  });
}

export function parseApiResultObject(obj: ApiResultObject): Result {
  switch (obj.type.class) {
  case 'boolean':
    if (obj.booleanValue === undefined) {
      throw 'missing booleanValue';
    }
    return {
      resultType: 'boolean',
      value: obj.booleanValue,
    }
  case 'number':
    if (obj.numberValue === undefined) {
      throw 'missing numberValue';
    }
    return {
      resultType: 'number',
      value: obj.numberValue,
    };
  case 'string':
    if (obj.stringValue === undefined) {
      throw 'missing stringValue';
    }
    return {
      resultType: 'string',
      value: obj.stringValue,
    };
  case 'lambda':
    if (obj.lambdaValue === undefined) {
      throw 'missing lambdaValue';
    }

    return {
      resultType: 'lambda',
      freeVariables: obj.lambdaValue.freeVariables,
    }
  case 'list':
    if (obj.listValue === undefined) {
      throw 'missing listValue';
    }
    return {
      resultType: 'list',
      elements: obj.listValue.elements.map(parseApiResultObject),
    }
  case 'record':
    if (obj.recordValue === undefined) {
      throw 'missing recordValue';
    }

    const properties = [];
    for (const [key, value] of Object.entries(obj.recordValue.properties)) {
      properties.push({
        name: key,
        value: parseApiResultObject(value),
      })
    }

    return {
      resultType: 'record',
      properties,
    }
  default:
    throw `unknown type ${obj.type.class}`;
  }
}

export interface ApiResult {
  error?: string,
  result?: ApiResultObject,
}

interface ApiResultObject {
  type: {
    class: string,
  },
  booleanValue?: boolean,
  numberValue?: number,
  stringValue?: string,
  lambdaValue?: {
    freeVariables: [string],
  },
  listValue?: {
    elements: ReadonlyArray<ApiResultObject>,
  },
  recordValue?: {
    properties: ReadonlyMap<string, ApiResultObject>,
  },
}
