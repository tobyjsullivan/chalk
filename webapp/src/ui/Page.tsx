import React, { Component, ComponentClass, StatelessComponent } from "react";
import {List} from 'immutable';
import {VariableState, PageState} from '../chalk/domain';

import './Page.css';
import FormulaWidget from "./FormulaWidget";

interface PropsType {
  currentPage: PageState,
  variables: List<VariableState>,
  onAddVar: () => any,
  onVarChange: (id: string, formula: string) => Promise<any>,
  onVarRename: (id: string, name: string) => Promise<any>,
}

const Page = ({onAddVar, variables, onVarChange, onVarRename}: PropsType) => {
  const widgets = variables.map(({id, name, formula, result}) => (
    <div key={name}>
      <FormulaWidget
        varName={name}
        formula={formula}
        result={result}
        onFormulaChange={(formula) => onVarChange(id, formula)}
        onNameChange={(name) => onVarRename(id, name)} />
    </div>
  ));

  return (
    <div className="Page">
      <a className="Page-link Page-addButton" onClick={() => onAddVar()} href="#">+ Add a block</a>
      <div>
        {widgets}
      </div>
    </div>
  );
};

export default Page;
