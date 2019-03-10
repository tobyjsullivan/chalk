import React, {Component} from 'react';
import enhanceWithClickOutside from 'react-click-outside';
import {Map} from 'immutable';

import {Result} from '../chalk/domain/resolver';
import ResultDisplay from './ResultDisplay';
import './FormulaWidget.css';

interface PropsType {
  varName: string;
  formula: string;
  result: Result;
  onFormulaChange: (formula: string) => Promise<any>;
  onNameChange: (name: string) => Promise<any>;
}

interface StateType {
  editing: boolean;
  editingName: boolean;
  disabled: boolean;
  currentFormula: string;
  currentName: string;
}

function isEndAction(e: React.KeyboardEvent) {
  // 13 => Enter
  // 27 => Escape
  if (e.keyCode == 13 || e.keyCode == 27) {
    e.preventDefault();
    return true;
  }
}

// Not a static component to allow the input focus and click-outside.
class FormulaWidget extends Component<PropsType, StateType> {
  input: HTMLInputElement | null = null;
  state = {
    editing: false,
    editingName: false,
    disabled: false,
    currentFormula: this.props.formula,
    currentName: this.props.varName,
  }

  componentDidMount() {
    this.focus()
  }

  componentDidUpdate() {
    this.focus()
  }

  focus() {
    this.input && this.state.editing ? this.input.focus() : {}
  }

  handleClickOutside() {
    const {editing} = this.state;

    if (editing) {
      this.endEditing();
    }
  }

  startEditingName() {
    this.setState({editingName: true});
  }

  endEditingName() {
    const {onNameChange} = this.props;
    const {editingName, currentName} = this.state;
    if (editingName) {
      this.setState({editingName: false});
      onNameChange(currentName);
    }
  }

  handleNameInputChanged(name: string) {
    this.setState({currentName: name});
  }

  startEditing() {
    this.setState({editing: true}, () => {
      this.focus();
    });
  }

  endEditing() {
    const {onFormulaChange} = this.props;
    const {editing, currentFormula: currentValue} = this.state;

    if (editing) {
      this.setState({disabled: true});
      onFormulaChange(currentValue).then(() => {
        this.setState({editing: false, disabled: false});
      });
    }
  }

  handleFormulaInputChanged(formula: string) {
    this.setState({currentFormula: cleanFormula(formula)});
  }

  render() {
    const {varName, formula, result} = this.props;
    const {editingName, editing, disabled} = this.state;
    let nameDisplay = (
      <div className="Widget-var_name" onClick={() => this.startEditingName()}>
        {varName}
      </div>
    );
    let valueDisplay = (
      <div className="Widget-result_window" onClick={() => this.startEditing()}>
        <ResultDisplay result={result} />
      </div>
    );
    if (editingName) {
      nameDisplay = (
        <div className="Widget-var_name">
          <input
            onChange={(e) => this.handleNameInputChanged(e.target.value)}
            onKeyDown={e => isEndAction(e) ? this.endEditingName() : {}}
            onBlur={() => this.endEditingName()}
            defaultValue={varName} />
        </div>
      );
    }
    if (editing) {
      valueDisplay = (
        <div className="Widget-formula_window">
          <input
            ref={(input) => {this.input = input}}
            className="Widget-formula_input"
            type="text"
            defaultValue={formula}
            disabled={disabled}
            onBlur={() => this.endEditing()}
            onKeyDown={e => isEndAction(e) ? this.endEditing() : {}}
            onChange={e => this.handleFormulaInputChanged(e.target.value)} />
        </div>
      );
    }

    return (
      <div className="Widget">
        {nameDisplay}
        {valueDisplay}
      </div>
    );
  }
}

const CHAR_REPLACEMENTS = Map({
  '“': '"',
  '”': '"',
});

function cleanFormula(formula: string): string {
  return CHAR_REPLACEMENTS.reduce((f, replacement, needle) => f.split(needle).join(replacement), formula);
}

export default enhanceWithClickOutside(FormulaWidget);
