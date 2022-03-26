import React, { useEffect, useState } from 'react';
import './application.scss';
import { observer } from 'mobx-react-lite';
import { useHistory } from 'react-router-dom';
import { APIRequestResult } from '@polo/common/api/common';
import { SessionSubscriptionEventType } from '@polo/common/state/models';
import { IApplication } from '@polo/common/state/models/application-model';
import { ISession } from '@polo/common/state/models/session-model';
import { ApplicationCheckouts } from './checkouts/application-checkouts';
import { ApplicationSessions } from './sessions/application-sessions';
import { ApplicationHeader } from './header/application-header';
import { Button } from '@polo/common/components/elements/button/button';
import { CubeIcon } from '@polo/common/components/elements/icons/cube/cube-icon';
import { useSubscription } from '@polo/common/state/models/subscription-hook';
import { NotificationType } from '@polo/common/state/models/notification-model';
import { useNotification } from '@polo/common/state/models/notification-hook';
import { TFailuresDictionary } from '@polo/common/state/models/failures-model';
import { buildFailedNotification } from '@polo/common/state/notifications/build-failed-notification';
import { useApplicationNotifications } from './notifications/application-notification-hook';
import { observe, values } from 'mobx';
import { IApplicationNotification } from '@polo/common/state/models/application-notification-model';
import { ApplicationNotifications } from './notifications/application-notifications';
import { onPatch } from 'mobx-state-tree';

type TProps = {
    sessions   : ISession[] | null;
    failures   : TFailuresDictionary | null;
    application: IApplication;

    moreThanOneApplication?: boolean;
    onApplicationsSelectorClick: () => void;
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
                notify({
                    text: 'Could not create new session.\n' + newSession.reason,
                    type: NotificationType.ERROR
                });
            }
        }
    }

    return <div className={`
        px-0
        mx-auto
        font-quicksand
        application`}>

        <ApplicationNotifications application={props.application} />
        <div className="hidden lg:block">
            <ApplicationHeader
                name={props.application.configuration.name}
                filename={props.application.filename}
                failures={props.failures} />
        </div>
        <div className="lg:hidden">
            <ApplicationHeader
                name={props.application.configuration.name}
                filename={props.application.filename}
                failures={props.failures}
                showApplicationSelector={props.moreThanOneApplication}
                onApplicationsSelectorClick={props.onApplicationsSelectorClick} />
        </div>

        
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