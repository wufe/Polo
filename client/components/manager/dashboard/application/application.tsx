import { APIRequestResult } from '@/api/common';
import { IApp, SessionSubscriptionEventType } from '@/state/models';
import { IApplication, IApplicationBranchModel } from '@/state/models/application-model';
import { ISession } from '@/state/models/session-model';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { ApplicationCheckouts } from './checkouts/application-checkouts';
import { ApplicationSessions } from './sessions/application-sessions';
import './application.scss';
import { DefaultModal } from '../../modal/default-modal';
import { useModal } from '../../modal/modal-hooks';
import { ApplicationOptionsModal } from './options/application-options-modal';
import { ApplicationHeader } from './header/application-header';
import { Button } from '@/components/shared/elements/button/button';
import { CubeIcon } from '@/components/shared/elements/icons/cube/cube-icon';
import { useSubscription } from '@/state/models/subscription-hook';
import { NotificationType } from '@/state/models/notification-model';
import { useNotification } from '@/state/models/notification-hook';
import { TFailuresDictionary } from '@/state/models/failures-model';
import { buildFailedNotification } from '@/state/notifications/build-failed-notification';

type TProps = {
    sessions   : ISession[] | null;
    failures   : TFailuresDictionary | null;
    application: IApplication;
}

export const Application = observer((props: TProps) => {

    const [newSessionCheckout, setNewSessionCheckout] = useState<string>("")
    const { subscribe } = useSubscription();
    const { notify } = useNotification();
    const history = useHistory();

    const onCheckoutChange = (value: string) => setNewSessionCheckout(value);

    const submitSessionCreation = async (checkout: string) => {
        if (!checkout) return;
        checkout = checkout.trim();
        if (checkout) {
            const newSession = await props.application.newSession(checkout);
            if (newSession.result === APIRequestResult.SUCCEEDED) {
                
                subscribe(newSession.payload.uuid, SessionSubscriptionEventType.FAIL, session => {
                    notify(buildFailedNotification(session, notification => {
                        notification.remove();
                        history.push(`/_polo_/session/failing/${session.uuid}`);
                    }));
                });
                
                history.push(`/_polo_/session/${newSession.payload.uuid}/`);
            } else {
                alert('Could not create new session.\n' + newSession.reason);
            }
        }
    }

    return <div className={`
        px-0
        mx-auto
        font-quicksand
        application`}>

        <ApplicationHeader
            name={props.application.configuration.name}
            filename={props.application.filename}
            failures={props.failures} />
        
        {props.sessions && props.sessions.length > 0 && <div className="py-4">
            <ApplicationSessions sessions={props.sessions} />
        </div>}

        {props.application.branchesMap.size > 0 && <div className="py-4">
            <ApplicationCheckouts
                branches={props.application.branchesMap}
                tags={props.application.tagsMap}
                onSessionCreationSubmission={submitSessionCreation} />
        </div>}
        
        <div className="mt-7 mb-0 flex justify-center">
            <div className="min-w-9/12 h-1 border-b border-gray-300 dark:border-gray-500"></div>
        </div>

        <div className="flex my-4 py-4 px-2 lg:px-6 flex-col">
            <span className="text-sm text-gray-500 opacity-80 mb-2">
                Or write down the commit you want to build.
            </span>
            <div className="flex items-center __input-container">
                <input
                    className=""
                    type="text"
                    placeholder="Commit, branch or tag"
                    value={newSessionCheckout}
                    onChange={e => onCheckoutChange(e.target.value)}
                    onKeyUp={e => e.key === 'Enter' && submitSessionCreation(newSessionCheckout)} />
                <Button
                    success
                    small
                    onClick={() => submitSessionCreation(newSessionCheckout)}
                    label="Create"
                    icon={<CubeIcon />} />
            </div>
        </div>
    </div>;
})