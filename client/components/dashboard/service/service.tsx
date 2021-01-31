import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { IService } from '@/state/models/service-model';
import { ISession } from '@/state/models/session-model';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { ServiceSessions } from './service-sessions';
import './service.scss';

type TProps = {
    sessions: ISession[] | null;
    service: IService;
}

export const Service = observer((props: TProps) => {

    const [newSessionCheckout, setNewSessionCheckout] = useState<string>("")
    const history = useHistory();

    const onCheckoutChange = (value: string) => setNewSessionCheckout(value);

    const submitSessionCreation = async () => {
        if (newSessionCheckout.trim()) {
            const newSession = await props.service.newSession(newSessionCheckout.trim());
            if (newSession.result === APIRequestResult.SUCCEEDED) {
                history.push(`/_polo_/session/${newSession.payload.uuid}`);
            } else {
                alert('Could not create new session.\n' + newSession.reason);
            }
        }
    }

    return <div className="
        p-6 divide-y dark:divide-gray-500
        mx-auto
        my-5 rounded-md shadow-lg max-w-7xl
        dark:bg-nord0">
            <div>
                <h3 className="text-xl font-normal leading-10 uppercase">{props.service.name}</h3>
                <div className="grid grid-cols-2 my-5 mt-2">
                    <div className="text-sm dark:text-gray-300">Remote:</div>
                    <div className="text-sm">{props.service.remote}</div>
                    <div className="text-sm dark:text-gray-300">Target:</div>
                    <div className="text-sm">{props.service.target}</div>
                </div>
            </div>
        
        {props.sessions && <div className="my-4">
            <ServiceSessions sessions={props.sessions} />
        </div>}
        
        <div>
            <div className="flex my-5 mb-0">
                <input className="flex-grow px-3 py-1 mx-3 text-sm border rounded-sm dark:bg-gray-300 dark:text-gray-700 dark:placeholder-gray-500 " type="text" placeholder="Checkout.." value={newSessionCheckout} onChange={e => onCheckoutChange(e.target.value)} onKeyUp={e => e.key === 'Enter' && submitSessionCreation()} />
                <button className="px-5 py-1 text-sm border rounded-sm hover:text-gray-50 dark:border-gray-500 hover:bg-blue-400 hover:border-blue-600" onClick={() => submitSessionCreation()}>Create</button>
            </div>
        </div>
    </div>;
})