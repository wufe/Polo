import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { SessionStatus } from '@/state/models/session-model';
import { observer } from 'mobx-react-lite';
import React, { useEffect } from 'react';
import { useHistory, useParams } from 'react-router-dom';
import { Session } from './session';

type TProps = {
    app: IApp;
}
export const SessionPage = observer((props: TProps) => {

    const { uuid } = useParams<{ uuid: string }>();
    const history = useHistory();

    const { session } = props.app;

    useEffect(() => {
        if (!session || session.uuid !== uuid) {
            props.app.retrieveSession(uuid)
                .then(request => {
                    if (request.result === APIRequestResult.FAILED) {
                        alert(request.reason);
                        history.push(`/_polo_/`);
                    }
                });
        }
        
    }, [uuid])

    useEffect(() => {

        if (session) {
            if (session.status === SessionStatus.STARTED) {
                session.track()
                    .then(request => {
                        if (request.result === APIRequestResult.SUCCEEDED) {
                            location.href = '/';
                        } else {
                            
                        }
                    });
            }
        }

    }, [session && session.status]);

    if (!session || session.uuid !== uuid) return null;

    return <Session session={session} />
});

export default SessionPage;