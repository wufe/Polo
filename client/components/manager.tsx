import { store } from '@/state/models';
import React from 'react';
import { BrowserRouter, Switch, Route } from 'react-router-dom';
import './manager.scss';

const Dashboard = React.lazy(() => import('@/components/dashboard/dashboard'));
const Session = React.lazy(() => import('@/components/session/session'));

export const ManagerApp = () => {
    return <div className="
        app__component
        text-black
        bg-gradient-to-br
        from-gray-50
        to-gray-100
        dark:from-nord-4
        dark:to-nord-1
        dark:bg-gray-800
        dark:text-gray-300">
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