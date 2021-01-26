import { store } from '@/state/models';
import React from 'react';
import { BrowserRouter, Switch, Route } from 'react-router-dom';
import './app.scss';

const Dashboard = React.lazy(() => import('@/components/dashboard/dashboard'));
const Session = React.lazy(() => import('@/components/session/session'));

export const App = () => {
    return <div className="app__component">
        <BrowserRouter>
            <React.Suspense fallback="">
                <Switch>
                    <Route path="/_polo_/" exact>
                        <Dashboard app={store.app} />
                    </Route>
                    <Route path="/_polo_/session/:uuid">
                        <Session app={store.app} />
                    </Route>
                </Switch>
            </React.Suspense>
        </BrowserRouter>
    </div>;
}