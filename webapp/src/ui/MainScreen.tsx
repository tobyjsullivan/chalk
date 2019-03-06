import React, { Component, ComponentClass, StatelessComponent } from "react";
import {List} from 'immutable';
import {VariableState, PageState} from '../chalk/domain';

import Header from './Header';
import Page from './Page';
import FormulaWidget from "./FormulaWidget";

import './MainScreen.css';

interface MainScreenProps {
  title: string,
  online: boolean | null,
  pages: List<PageState>,
  currentPage: PageState,
  variables: List<VariableState>,
  onAddVar: () => any,
  onOpenPage: (pageId: string) => any,
  onVarChange: (id: string, formula: string) => Promise<any>,
  onVarRename: (id: string, name: string) => Promise<any>,
}

const MainScreen = ({title, online, currentPage, pages, onAddVar, variables, onOpenPage, onVarChange, onVarRename}: MainScreenProps) => {
  let status = 'checking...';
  if (online !== null) {
    status = online ? 'online' : 'connection error';
  }

  const pageMenuItems = pages.map(({id}) => (
    <li key={id}>
      <a
        href="#"
        onClick={() => onOpenPage(id)}>
        {id}
      </a>
    </li>
  ));
  const pageMenu = (
    <ul>
      {pageMenuItems}
    </ul>
  );

  return (
    <div className="MainScreen">
      <Header title={title} />
      {pageMenu}
      <Page
        currentPage={currentPage}
        variables={variables}
        onAddVar={onAddVar}
        onVarChange={onVarChange}
        onVarRename={onVarRename} />
      <p>Status: {status}</p>
    </div>
  );
};

export default MainScreen;
