import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { IApplication, IApplicationBranchModel } from '@/state/models/application-model';
import { ISession } from '@/state/models/session-model';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { ApplicationBranches } from './application-branches';
import { ApplicationSessions } from './application-sessions';
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
        divide-y
        divide-gray-200
        dark:divide-gray-600
        mx-auto
        my-5 rounded-md shadow-lg
        bg-gray-50
        dark:bg-nord0
        font-quicksand
        max-w-5xl
        ${!props.isOpen ? ' max-h-20 overflow-hidden dark:hover:bg-nord3' : ''}`}
        >
        <div className={`flex flex-col lg:flex-row items-stretch lg:items-center justify-between relative lg:-mx-6 px-2 lg:px-12 lg:pr-12 ${props.toggleEnabled ? 'cursor-pointer' : ''}`} onClick={props.onToggle}>
            <h3 className="h-20 flex justify-center lg:justify-start items-center text-mg font-normal leading-10 uppercase col-span-3 overflow-hidden overflow-ellipsis whitespace-nowrap text-center lg:text-left" title={props.application.configuration.name}>{props.application.configuration.name}</h3>
            <div className="col-span-4 lg:px-10 hidden lg:block">
                <div className="text-xs text-gray-500 uppercase my-3 lg:my-0">Remote:</div>
                <div
                    className="text-sm overflow-hidden overflow-ellipsis whitespace-nowrap"
                    title={props.application.configuration.remote}>{props.application.configuration.remote}</div>
            </div>
            {props.isOpen && props.toggleEnabled && <svg width={16} height={16} className="hidden lg:block absolute right-10 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>}
            {!props.isOpen && props.toggleEnabled && <svg width={16} height={16} className="hidden lg:block absolute right-10 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>}
        </div>

        <div className="py-4 px-2 lg:px-6 hidden lg:block">
            <div className="col-span-2">
                <h4 className="my-1 text-xs text-gray-500 uppercase">Target:</h4>
                <div
                    className="text-sm overflow-hidden overflow-ellipsis whitespace-nowrap"
                    title={props.application.configuration.target}>{props.application.configuration.target}</div>
            </div>
        </div>
        
        {props.sessions && props.sessions.length > 0 && <div className="py-4 px-2 lg:px-6 bg-gradient-to-bl from-nord4 to-white dark:from-nord-1 dark:to-nord-4">
            <ApplicationSessions sessions={props.sessions} />
        </div>}

        {props.application.branchesMap.size > 0 && <div className="py-4">
            <ApplicationBranches branches={props.application.branchesMap} onSessionCreationSubmission={submitSessionCreation} />
        </div>}
        
        <div className="flex my-4 py-4 px-2 lg:px-6">
            <input
                className="flex-grow px-1 lg:px-3 py-1 mr-3 text-sm border rounded-sm dark:bg-gray-300 dark:text-gray-700 dark:placeholder-gray-500"
                type="text"
                placeholder="Checkout a commit, a branch or a tag.."
                value={newSessionCheckout}
                onChange={e => onCheckoutChange(e.target.value)}
                onKeyUp={e => e.key === 'Enter' && submitSessionCreation(newSessionCheckout)} />
            <button className="px-5 py-1 text-sm border rounded-sm hover:text-gray-50 dark:border-gray-500 hover:bg-blue-400 hover:border-blue-600" onClick={e => submitSessionCreation(newSessionCheckout)}>Create</button>
        </div>
    </div>;
})