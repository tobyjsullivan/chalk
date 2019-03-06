import React, { Component, StatelessComponent, ComponentClass } from 'react';
import {List, Map} from 'immutable';
import {VariableState, PageState} from './chalk/domain';
import { SessionState } from './chalk/domain';
import MainScreen from './ui/MainScreen';
import LoadingScreen from './ui/LoadingScreen';
import './App.css';

const DEFAULT_FORMULA = '"Tap Here"';
const PAGE_PATH_REGEXP = /^\/([a-fA-F0-9-]+)$/;

interface AppProps {
  currentPath: string,
  checkConnection: () => Promise<any>,
  getSession: () => Promise<SessionState>,
  getPageVariables: (pageId: string) => Promise<List<VariableState>>,
  createVariable: (pageId: string, varName: string, formula: string) => Promise<VariableState>,
  updateVariable: (id: string, formula: string) => Promise<VariableState>,
  renameVariable: (id: string, name: string) => Promise<VariableState>,
}

interface AppState {
  online: boolean | null,
  currentPage: PageState | null,
  nextVarNum: number,
  session: SessionState | null,
  currentPageVars: Map<string, VariableState>,
}

class App extends Component<AppProps, AppState> {
  constructor(props: AppProps) {
    super(props);
  }

  state: AppState = {
    online: null,
    currentPage: null,
    nextVarNum: 1,
    session: null,
    currentPageVars: Map<string, VariableState>(),
  }

  async initSession() {
    const {getSession, currentPath} = this.props;
    const session = await getSession();
    this.setState({session});

    if (PAGE_PATH_REGEXP.test(currentPath)) {
      const match = PAGE_PATH_REGEXP.exec(currentPath);
      if (match === null) {
        alert('page not found');
        this.openDefaultPage();
        return;
      }

      const pageId = match[1];
      this.openPage(pageId)
    } else {
      this.openDefaultPage();
    }
  }

  openPage(pageId: string) {
    // TODO (toby): Fetch actual page
    const page = {
      id: pageId,
    };

    this.handlePageOpened(page);
  }

  openDefaultPage() {
    const {session} = this.state;
    if (session === null) {
      return;
    }

    const defaultPage = session.pages.first(null);
    if (defaultPage === null) {
      return;
    }

    this.handlePageOpened(defaultPage);
  }

  componentDidMount() {
    const {checkConnection} = this.props;

    checkConnection().then(() => {
      this.setState({online: true});
    }).catch(() => {
      this.setState({online: false});
    });

    // We initialise sessions within the App component because we don't want to block first render.
    this.initSession();
  }

  async getNextVarName(): Promise<string> {
    return new Promise((res) => {
      this.setState(({nextVarNum}) => {
        const varName = `var${nextVarNum}`;

        res(varName);
        return {
          nextVarNum: nextVarNum + 1,
        };
      });
    });
  }

  async onAdd() {
    console.log('onAdd() executing...');
    const {createVariable} = this.props;
    const {currentPage} = this.state;
    if (currentPage === null) {
      console.warn('var added before page loaded.');
      return;
    }

    const varName = await this.getNextVarName();
    createVariable(currentPage.id, varName, DEFAULT_FORMULA).then((varState) => {
      this.reloadPageVars();
    });
  }

  async handleVarChanged(varId: string, formula: string): Promise<void> {
    await this.props.updateVariable(varId, formula);

    this.reloadPageVars();
  }

  async handleVarRenamed(varId: string, name: string): Promise<void> {
    try {
      await this.props.renameVariable(varId, name);

      this.reloadPageVars();
    } catch (e) {
      console.error(e);
      alert('error renaming var: ' + e);
    }
  }

  async handlePageOpened(page: PageState) {
    this.setState({currentPage: page}, () => {
      this.reloadPageVars();
    });
  }

  async reloadPageVars() {
    // Load page variables
    const {getPageVariables} = this.props;
    const {currentPage} = this.state;

    if (currentPage === null) {
      return;
    }

    const vars = await getPageVariables(currentPage.id);
    const varMap = vars.toMap().mapKeys((_, v) => v.id);
    this.setState({currentPageVars: varMap});
  }

  render() {
    const {online, currentPage, currentPageVars} = this.state;

    const title = "Messy";

    let currentScreen;
    if (currentPage === null) {
      currentScreen = (<LoadingScreen title={title} />);
    } else {
      currentScreen = (
        <MainScreen
          title={title}
          online={online}
          currentPage={currentPage}
          pages={List()}
          variables={List(currentPageVars.values())}
          onAddVar={() => this.onAdd()}
          onOpenPage={(pageId) => this.openPage(pageId)}
          onVarChange={(id, formula) => this.handleVarChanged(id, formula)}
          onVarRename={(id, name) => this.handleVarRenamed(id, name)} />
      );
    }

    return (
      <div className="App">
        {currentScreen}
      </div>
    );
  }
}

export default App;
