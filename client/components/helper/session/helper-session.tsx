import { APIRequestResult } from '@/api/common';
import { IAPISession, retrieveSessionAgeAPI, untrackSessionAPI } from '@/api/session';
import React, { memo, useEffect, useRef, useState } from 'react';
import { HelperStatus, HelperStatusContext } from '../contexts';
import { expiredAgeValue, noExpirationAgeValue } from '../status/helper-status-provider';
import './helper-session.scss';



export const SessionMaxAge = () => {
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

    const maxAgeVisible =
        props.session.maxAge > noExpirationAgeValue &&
        props.session.maxAge > expiredAgeValue;

    return <div className="helper-session__component">
        <div>On checkout <b>{props.session.checkout}</b></div>
        <div><small>Session: {props.session.uuid.split('-')[0]}</small></div>
        {maxAgeVisible && <div>
            <SessionMaxAge />
        </div>}
        <div className="__detach" onClick={detach}>Exit</div>
    </div>
}