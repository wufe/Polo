import { APIRequestResult } from '@/api/common';
import { ISession } from '@/state/models';
import React from 'react';

type TProps = {
    sessions: ISession[] | null;
}

export const ServiceSessions = (props: TProps) => {

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
        <h4 className="my-3">Sessions:</h4>
        {props.sessions.map((session, key) =>
            <div
                key={key}
                className="flex justify-between">
                <span>Session: {session.uuid}</span>
                <span>
                    <span className="text-sm underline cursor-pointer inline-block mx-3" onClick={() => attachToSession(session)}>Attach</span>
                    <span className="text-sm underline cursor-pointer inline-block mx-3" onClick={() => killSession(session)}>Kill</span>
                </span>
            </div>)}
    </>;
}