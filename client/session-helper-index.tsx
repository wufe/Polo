import { cast, getSnapshot, getType } from 'mobx-state-tree';
import React, { createContext } from 'react';
import { render } from 'react-dom';
import { IAPISession } from './api/session';
import { SessionHelperApp } from './components/session-helper';
import { AppModel, castAPISessionToSessionModel, createStore, initialAppState, ISession } from './state/models';

const store = createStore({
    app: AppModel.create({
        session: castAPISessionToSessionModel(window.currentSession) as any
    })
});
const context = createContext(store);

render(<SessionHelperApp store={store} />, document.getElementById('polo-session-helper'));

declare global {
    interface Window {
        currentSession: IAPISession;
    }
}