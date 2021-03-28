import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { IApplication, IApplicationBranchModel } from '@/state/models/application-model';
import { ISession } from '@/state/models/session-model';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { ApplicationCheckouts } from './checkouts/application-checkouts';
import { ApplicationSessions } from './sessions/application-sessions';
import './application.scss';

type TProps = {
    isOpen       : boolean;
    onToggle     : () => void;
    toggleEnabled: boolean;
    sessions     : ISession[] | null;
    application  : IApplication;
}

export const Application = observer((props: TProps) => {

    const [newSessionCheckout, setNewSessionCheckout] = useState<string>("")
    const history = useHistory();

    const onCheckoutChange = (value: string) => setNewSessionCheckout(value);

    const submitSessionCreation = async (checkout: string) => {
        if (!checkout) return;
        checkout = checkout.trim();
        if (checkout) {
            const newSession = await props.application.newSession(checkout);
            if (newSession.result === APIRequestResult.SUCCEEDED) {
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
        <div>
            <h3 className="text-xl lg:text-2xl leading-5 font-bold overflow-hidden overflow-ellipsis whitespace-nowrap" title={props.application.configuration.name}>{props.application.configuration.name}</h3>
            <span className="text-gray-400 text-sm">{props.application.filename}</span>
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
                <div className="__button --success --small" onClick={e => submitSessionCreation(newSessionCheckout)}>
                    <span>Create</span>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                </div>
            </div>
        </div>
    </div>;
})