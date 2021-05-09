import { store } from '@/state/models';
import React from 'react';
import { BrowserRouter, Switch, Route, useRouteMatch } from 'react-router-dom';
import './index.scss';
import whiteLogo from '@/assets/white-logo.png';
import blackLogo from '@/assets/black-logo.png';
import { observer } from 'mobx-react-lite';
import FailingSessionPage from './session/failing-session-page';
import { Notifications } from './notifications/notifications';

const Dashboard = React.lazy(() => import('@/components/manager/dashboard/dashboard-page'));
const Session = React.lazy(() => import('@/components/manager/session/session-page'));



export const ManagerApp = observer(() => {
    return <>
            <div className={`
            flex-1
            w-full
            flex
            flex-col
            items-stretch
            justify-stretch
            min-w-0
            min-h-0
            relative
            text-black
            bg-gradient-to-br
            from-gray-50
            to-gray-100
            dark:from-nord-4 dark:to-nord-1
            dark:bg-gray-800
            dark:text-gray-300
            manager-app
            ${store.app.modal.visible ? '--blurred' : ''}
            `}>
            <div className="flex pt-10 pb-8 z-10">
                <div className="w-full mx-auto lg:max-w-1500 lg:px-20 text-center lg:text-left">
                    <a href="/_polo_/">
                        <picture>
                            <source srcSet={whiteLogo} media="(prefers-color-scheme: dark)" />
                            <img src={blackLogo} width="200" className="cursor-pointer inline-block" />
                        </picture>
                    </a>
                </div>
            </div>
            <BrowserRouter>
                <React.Suspense fallback="">
                    <Switch>
                        <Route path="/_polo_/" exact>
                            <Dashboard app={store.app} />
                        </Route>
                        <Route path="/_polo_/session/failing/:uuid">
                            <FailingSessionPage app={store.app} />
                        </Route>
                        <Route path="/_polo_/session/:uuid">
                            <SessionRoute />
                        </Route>
                    </Switch>
                </React.Suspense>
            </BrowserRouter>
            <div className="absolute right-5 bottom-5 z-10 uppercase text-xs tracking-widest flex items-center">
                <a className="" target="_blank" href="https://github.com/wufe/polo">
                    Powered by <b>@wufe/polo</b>
                </a>
            </div>
        </div>
        <Notifications />
    </>;
});

function SessionRoute() {

    const { path } = useRouteMatch();

    return <Switch>
        <Route path={`${path}/*`}>
            <Session app={store.app} />
        </Route>
    </Switch>
}

