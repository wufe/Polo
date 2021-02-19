import { APIRequestResult } from '@/api/common';
import { IAPISession, retrieveSessionAgeAPI, untrackSessionAPI } from '@/api/session';
import React, { memo, useEffect, useRef, useState } from 'react';
import { HelperStatus, HelperStatusContext } from '../contexts';
import './helper-session.scss';



export const SessionMaxAge = (props: { age: number, uuid: string }) => {
    

    return <HelperStatusContext.Consumer>
        {({age}) => <>Expires in: {age}s</>}
    </HelperStatusContext.Consumer>
}

type TProps = {
    session: IAPISession;
}

export const HelperSession = (props: TProps) => {

    const detach = async () => {
        const untrack = await untrackSessionAPI();
        if (untrack.result === APIRequestResult.SUCCEEDED) {
            location.href = '/';
        }
    }

    return <div className="helper-session__component">
        Session: {props.session.uuid}
        <br />
        <SessionMaxAge
            age={props.session.maxAge}
            uuid={props.session.uuid} />
        <br />
        <span className="__detach" onClick={detach}>Detach</span>
    </div>
}