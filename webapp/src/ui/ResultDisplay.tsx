import React from 'react';
import {Result} from '../chalk/domain/resolver';
import './ResultDisplay.css';

interface ResultDisplayCellPropsType {
  value: string,
}

const ResultDisplayCell = ({value}: ResultDisplayCellPropsType) => {
  return (<td className="ResultDisplay-cell">{value}</td>);
};

interface ResultDisplayRowPropTypes {
  values: ReadonlyArray<string>,
}

const ResultDisplayRow = ({values}: ResultDisplayRowPropTypes) => {
  const cells = values.map((v) => (<ResultDisplayCell key={v} value={v}/>));

  return (
      <tr className="ResultDisplay-row">
        {cells}
      </tr>
  );
};

interface ResultDisplayTablePropTypes {
  rows: ReadonlyArray<ReadonlyArray<string>>,
}

const ResultDisplayTable = ({rows}: ResultDisplayTablePropTypes) => {
  const rowElements = rows.map((r) => (<ResultDisplayRow  key={r.join(',')} values={r} />));

  return (
      <table className="ResultDisplay-table">
        <tbody>
          {rowElements}
        </tbody>
      </table>
  );
};

const SingleCell = ({content}: {content: string}) => {
  return (<ResultDisplayTable rows={[[content]]}/>)
};

interface ResultDisplayPropsType {
  result: Result,
}

const ResultDisplay = ({result}: ResultDisplayPropsType) => {
  let content: JSX.Element;

  switch (result.resultType) {
    case 'none':
      content = (<SingleCell content="" />);
      break;
    case 'boolean':
      content = (<SingleCell content={result.value ? 'TRUE' : 'FALSE'} />);
      break;
    case 'lambda':
      content = (<SingleCell content={`λ (${result.freeVariables.join(', ')})`} />);
      break;
    case 'list':
      const items = result.elements.map((res, i) => (
        <li key={i} className="ResultDisplay-listItem">
          <ResultDisplay result={res} />
        </li>
      ));
      content = (
        <ul className="ResultDisplay-list">
          {items}
        </ul>
      );
      break;
    case 'number':
      content = (<SingleCell content={`${result.value}`} />);
      break;
    case 'record':
      const propRows = result.properties.map(({name, value}) => (
        <tr className="ResultDisplay-recordRow" key={name}>
          <th className="ResultDisplay-recordProperty">{name}</th>
          <td className="ResultDisplay-recordValue"><ResultDisplay result={value} /></td>
        </tr>
      ));

      content = (
        <table className="ResultDisplay-record">
          <tbody>
            {propRows}
          </tbody>
        </table>
      );
      break;
    case 'string':
      content = (<SingleCell content={result.value} />);
      break;
    case 'error':
      content = (
        <p className="ResultDisplay-error">{result.message}</p>
      );
      break;
    default:
        throw 'Unexpected type: ' + result;
  }

  return (
    <div className="ResultDisplay">
      {content}
    </div>
  );
};

export default ResultDisplay;
