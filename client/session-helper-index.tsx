import { getSnapshot } from 'mobx-state-tree';
import React, { createContext } from 'react';
import { render } from 'react-dom';
import { SessionHelperApp } from './components/session-helper';
import { AppModel, createStore, initialAppState, ISession } from './state/models';

const store = createStore({
    app: AppModel.create({
        session: window.currentSession
    })
});
const context = createContext(store);

render(<SessionHelperApp store={store} />, document.getElementById('polo-session-helper'));

declare global {
    interface Window {
        currentSession: ISession;
    }
}