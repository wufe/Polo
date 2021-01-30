import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { IService } from '@/state/models/service-model';
import { ISession } from '@/state/models/session-model';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import './service.scss';

type TProps = {
    app: IApp;
    service: IService;
}

export const Service = observer((props: TProps) => {

    const [sessions, setSessions] = useState<ISession[]>([]);
    const [newSessionCheckout, setNewSessionCheckout] = useState<string>("")
    const history = useHistory();

    useEffect(() => {
        setSessions(props.app.sessionsByServiceName[props.service.name] || []);
    }, [props.app.sessionsByServiceName]);

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

    return <div className="p-6 mx-auto my-5 rounded-md shadow-lg max-w-7xl
        dark:bg-nord0">
        <h3 className="text-xl font-normal leading-10 uppercase">{props.service.name}</h3>
        <div className="grid grid-cols-2 my-5 mt-2">
            <div className="text-sm dark:text-gray-300">Remote:</div>
            <div className="text-sm">{props.service.remote}</div>
            <div className="text-sm dark:text-gray-300">Target:</div>
            <div className="text-sm">{props.service.target}</div>
            {/* <div className="text-sm dark:text-gray-300">Branches:</div>
            <div className="text-sm">
                {props.service.branches.map((branch, key) => <span key={key} className={"leading-8"}>
                    {branch}
                </span>)}
            </div> */}
        </div>
        
        {!!sessions.length && <>
            <hr />
            <br />
            <h4>Sessions:</h4>
            <br />
            {sessions.map((session, key) => <div key={key}>Session: {session.uuid}</div>)}
        </>}
        <>
            <div className="flex my-5 mb-0">
                <input className="flex-grow px-3 py-1 mx-3 text-sm border rounded-sm dark:bg-gray-300 dark:text-gray-700 dark:placeholder-gray-500 " type="text" placeholder="Checkout.." value={newSessionCheckout} onChange={e => onCheckoutChange(e.target.value)} onKeyUp={e => e.key === 'Enter' && submitSessionCreation()} />
                <button className="px-5 py-1 text-sm border rounded-sm hover:text-gray-50 dark:border-gray-500 hover:bg-blue-400 hover:border-blue-600" onClick={() => submitSessionCreation()}>Create</button>
            </div>
        </>
    </div>;
})