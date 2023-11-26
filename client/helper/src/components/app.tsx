import { IAPISession } from '@polo/common/api/session';
import React, {  } from 'react';
import './app.scss';
import { HelperSession } from './session/helper-session';
import { render } from 'react-dom';
import { HelperStatusContext } from './contexts';
import { HelperOverlay } from './overlay/helper-overlay';
import { HelperStatusProvider } from './status/helper-status-provider';

type TProps = {
    session: IAPISession;
}
export const App = (props: TProps) => {

    if (!props.session)
        return null;

    const { age, status, killReason, replacedBy } = props.session;

    return <HelperStatusProvider
        uuid={props.session.uuid}
        initial={{
            age,
            status,
            killReason,
            replacedBy
        }}>
        <HelperStatusContext.Consumer>
            {({ helperStatus }) => <HelperOverlay helperStatus={helperStatus} />}
        </HelperStatusContext.Consumer>
        <div className="session-helper__component">
            <HelperSession session={props.session} />
        </div>
    </HelperStatusProvider>
};

render(<App session={window.currentSession} />, document.getElementById('polo-session-helper'));

setInterval(() =>
{
    if (!document.getElementById('polo-session-helper'))
    {
        const helper = document.createElement('div');
        helper.id = 'polo-session-helper';
        helper.setAttribute('data-pos-x', 'left');
        helper.setAttribute('data-pos-y', 'bottom');
        document.body.prepend(helper);

        render(<App session={window.currentSession} />, document.getElementById('polo-session-helper'));
    }
}, 1000);


declare global {
    interface Window {
        currentSession: any;
    }
}