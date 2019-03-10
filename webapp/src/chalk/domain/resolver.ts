export type Result = Error | NoneType | Boolean | Lambda | List | Number | Record | String;

export const None: NoneType = {
  resultType: 'none',
}

interface NoneType {
  resultType: 'none',
}

export interface Boolean {
  resultType: 'boolean',
  value: boolean,
}

export interface Lambda {
  resultType: 'lambda',
  freeVariables: [string],
}

export interface List {
  resultType: 'list',
  elements: ReadonlyArray<Result>,
}

export interface Number {
  resultType: 'number',
  value: number,
}

export interface Record {
  resultType: 'record',
  properties: ReadonlyArray<RecordProperty>;
}

export interface RecordProperty {
  name: string;
  value: Result;
}

export interface String {
  resultType: 'string',
  value: string,
}

interface Error {
  resultType: 'error',
  message: string,
}
