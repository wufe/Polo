import { IApp } from '@/state/models/app-model';
import { useNotification } from '@/state/models/notification-hook';
import { ISession, ISessionLog } from '@/state/models/session-model';
import { buildFailedNotification } from '@/state/notifications/build-failed-notification';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React, { useState } from 'react';
import { useHistory } from 'react-router-dom';
import { CommitMessage } from '../shared/commit-message';
import { SessionLogs } from './session-logs';
import { useSessionRetrieval } from './session-retrieval-hook';

type TProps = {
    app: IApp;
    session: ISession;
}
export const Session = observer((props: TProps) => {

    const [overlayBottom, setOverlayBottom] = useState(100);
    const history = useHistory();
    const { notify } = useNotification();

    const onSessionFail = () => {
        notify(
            buildFailedNotification(
                props.session,
                notification =>
                    notification.remove()
                )
            );
        history.replace(`/_polo_/session/failing/${props.session.uuid}`);
    }
    
    useSessionRetrieval(props.app.failures.retrieveFailedSession, onSessionFail, props.session);

    const setOverlayProportions = (proportions: number) => {
        const percentage = parseInt(`${proportions * 100}`);
        const inversePercentage = 100 - percentage;
        setOverlayBottom(inversePercentage);
    }

    return <div className="
        mx-auto w-full max-w-6xl flex flex-col min-w-0 min-h-0 flex-1 pt-3 font-quicksand" style={{height:'calc(100vh - 120px)'}}>
        <div className="main-gradient-faded absolute left-0 right-0 top-0 pointer-events-none" style={{ bottom: `${overlayBottom}%`, zIndex: 1 }}></div>
        <h1 className="text-4xl px-2 lg:px-0 mb-3 font-quicksand font-light text-nord1 dark:text-nord5 z-10">
            Session
        </h1>
        <div className="text-lg text-nord1 dark:text-nord5 mb-4 z-10 border-l pl-3 border-gray-500">
            <span>{props.session.displayName}</span>
        </div>
        <CommitMessage {...props.session} maxHeight />
        <SessionLogs
            logs={values(props.session.logs) as any as ISessionLog[]}
            onLogsProportionChanged={setOverlayProportions} />
    </div>
});

export default Session;