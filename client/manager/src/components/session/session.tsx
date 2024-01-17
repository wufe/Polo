import { SessionLogsContainer } from '@/components/session/session-logs-container';
import { SessionTerminalContainer } from '@/components/session/session-terminal-container';
import { IApp } from '@polo/common/state/models/app-model';
import { useNotification } from '@polo/common/state/models/notification-hook';
import { ISession } from '@polo/common/state/models/session-model';
import { buildFailedNotification } from '@polo/common/state/notifications/build-failed-notification';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { CommitMessage } from '../shared/commit-message';
import 'xterm/css/xterm.css';
import '@polo/manager/src/components/session/session.scss';
import { SessionIntegrationsStatus } from './integrations-status/session-integrations-status';

declare global {
    interface Window {
        configuration: {
            advancedTerminalOutput: boolean;
            integrationsPublicURL: string;
        };
    }
}

const useAdvancedTerminal = window.configuration.advancedTerminalOutput;

type TProps = {
    app: IApp;
    session: ISession;
};

export const Session = observer((props: TProps) => {
    const history = useHistory();
    const { notify } = useNotification();
    const [overlayBottom, setOverlayBottom] = useState(100);

    const onSessionFail = () => {
        notify(
            buildFailedNotification(
                props.session,
                notification =>
                    notification.remove()
            )
        );
        history.replace(`/_polo_/session/failing/${props.session.uuid}`);
    };

    return <div className="
        mx-auto w-full max-w-6xl flex flex-col min-w-0 min-h-0 flex-1 pt-3 font-quicksand" style={{ maxHeight: 'calc(100vh - 120px)' }}>
        <div className="main-gradient-faded absolute left-0 right-0 top-0 pointer-events-none" style={{ bottom: `${overlayBottom}%`, zIndex: 1 }}></div>
        <h1 className="text-4xl px-2 lg:px-0 mb-3 font-quicksand font-light text-nord1 dark:text-nord5 z-10">
            Session
        </h1>
        <div className="text-lg text-nord1 dark:text-nord5 mb-4 z-10 border-l pl-3 border-gray-500 flex justify-between min-w-full">
            <span>{props.session.displayName}</span>
            <div>
                <SessionIntegrationsStatus
                    session={props.session}
                    integrationsStatus={props.session.integrations} />
            </div>
        </div>
        <CommitMessage {...props.session} maxHeight />
        {useAdvancedTerminal && <SessionTerminalContainer app={props.app} session={props.session} onSessionFail={onSessionFail} />}
        {!useAdvancedTerminal && <SessionLogsContainer app={props.app} session={props.session} setOverlayBottom={setOverlayBottom} onSessionFail={onSessionFail} />}

    </div>
});

export default Session;