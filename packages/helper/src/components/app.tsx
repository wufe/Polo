import {render} from 'preact';
import { IAPISession } from '@polo/common/api/session';
import './app.scss';
import { HelperSession } from './session/helper-session';
import { HelperStatusContext } from './contexts';
import { HelperOverlay } from './overlay/helper-overlay';
import { HelperStatusProvider } from './status/helper-status-provider';

type TProps = {
    session: IAPISession;
}
export const App = (props: TProps) => {

    if (!props.session)
        return <pre>No session</pre>;

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

declare global {
    interface Window {
        currentSession: any;
    }
}