import { store } from '@/state/models';
import React from 'react';
import { BrowserRouter, Switch, Route } from 'react-router-dom';
import './index.scss';

const Dashboard = React.lazy(() => import('@/components/manager/dashboard/dashboard-page'));
const Session = React.lazy(() => import('@/components/manager/session/session-page'));

export const ManagerApp = () => {
    return <div className="
        flex-1
        w-full
        flex
        items-stretch
        justify-stretch
        min-w-0
        min-h-0
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