import { IAPISession } from '@/api/session';
import React, {  } from 'react';
import './index.scss';
import { HelperSession } from './session/helper-session';
import { render } from 'react-dom';
import { HelperStatusContext } from './contexts';
import { HelperOverlay } from './overlay/helper-overlay';
import { HelperStatusProvider } from './status/helper-status-provider';

type TProps = {
    session: IAPISession;
}
export const HelperApp = (props: TProps) => {

    if (!props.session)
        return null;

    return <HelperStatusProvider
        uuid={props.session.uuid}
        maxAge={props.session.maxAge}
        age={props.session.maxAge}>
        <HelperStatusContext.Consumer>
            {({ status }) => <HelperOverlay status={status} />}
        </HelperStatusContext.Consumer>
        <div className="session-helper__component">
            <HelperSession session={props.session} />
        </div>
    </HelperStatusProvider>
};

render(<HelperApp session={window.currentSession} />, document.getElementById('polo-session-helper'));

declare global {
    interface Window {
        currentSession: any;
    }
}