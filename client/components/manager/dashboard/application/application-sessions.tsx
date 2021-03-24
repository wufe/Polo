import { APIRequestResult } from '@/api/common';
import { ISession } from '@/state/models';
import { SessionStatus } from '@/state/models/session-model-enums';
import { observer } from 'mobx-react-lite';
import React from 'react';
import loading from '../../../../assets/loading.svg';
import './application-sessions.scss';

export const noExpirationAgeValue = -1;

type TProps = {
    sessions: ISession[];
}

export const ApplicationSessions = observer((props: TProps) => {

    const attachToSession = async (session: ISession) => {
        const track = await session.track();
        if (track.result === APIRequestResult.SUCCEEDED) {
            location.href = '/';
        }
    }
    const killSession = async (session: ISession) => {
        if (confirm(`You are going to delete the session. Are you sure?`)) {
            await session.kill();
        }
    }

    return <>
        <h4 className="my-1 text-xs text-gray-500 uppercase">Sessions:</h4>
        {props.sessions
            .filter(session => !session.replacesSession)
            .map((session, key) =>
            <div
                key={key}
                className="flex flex-col lg:flex-row lg:items-center py-1 my-5 lg:my-0
                    application-sessions">
                <span className="flex flex-nowrap lg:w-24">
                    <span className="w-5 text-center inline-block">
                        {
                            (
                                session.status === SessionStatus.STARTING ||
                                session.beingReplaced
                            ) &&
                            <img src={loading} width={12} height={12} className="mr-1" />}
                    </span>
                    <span className="text-xs uppercase font-bold" style={{ color: colorByStatus(session.status) }}>{session.status}</span>
                </span>
                <span className="inline-block py-3 lg:py:0 text-center lg:text-left text-sm mr-3 flex-grow px-2 lg:px-10 flex-1 whitespace-nowrap overflow-hidden overflow-ellipsis">{session.checkout}</span>
                {session.age > noExpirationAgeValue && <span className="text-xs uppercase text-gray-500 px-2 lg:px-10">
                    <span className="hidden lg:inline-block pr-1">Expires in </span><span>{session.age}s</span>
                </span>}
                <span className="lg:px-10 text-center whitespace-nowrap my-3 lg:my-0">
                    <span className="__button --success border-nord14 " onClick={() => attachToSession(session)}>Enter</span>
                    <span className="__button --danger" onClick={() => killSession(session)}>Delete</span>
                </span>
            </div>)}
    </>;
});

function colorByStatus(status: SessionStatus): string {
    switch (status) {
        case SessionStatus.STARTED:
            return '#a3be8c';
        case SessionStatus.STARTING:
        case SessionStatus.DEGRADED:
            return '#ebcb8b';
        case SessionStatus.STOPPING:
            return '#d08770';
        case SessionStatus.START_FAILED:
            return '#bf616a';
    }
}