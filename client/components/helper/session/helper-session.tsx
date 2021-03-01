import { APIRequestResult } from '@/api/common';
import { IAPISession, retrieveSessionAgeAPI, untrackSessionAPI } from '@/api/session';
import React, { memo, useEffect, useRef, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { HelperStatus, HelperStatusContext } from '../contexts';
import { expiredAgeValue, noExpirationAgeValue } from '../status/helper-status-provider';
import './helper-session.scss';
import { useClipboard } from './use-clipboard';

export const SessionMaxAge = () => {
    return <HelperStatusContext.Consumer>
        {({ age }) => <>Expires in <b>{age}s</b></>}
    </HelperStatusContext.Consumer>
}

type TProps = {
    session: IAPISession;
}

export const HelperSession = (props: TProps) => {

    const [open, setOpen] = useState(false);
    const copy = useClipboard();

    const detach = async () => {
        const untrack = await untrackSessionAPI();
        if (untrack.result === APIRequestResult.SUCCEEDED) {
            location.href = '/_polo_/';
        }
    }

    const maxAgeVisible =
        props.session.maxAge > noExpirationAgeValue &&
        props.session.maxAge > expiredAgeValue;

    const copyPermalink = () => {
        let pathname = location.pathname;
        pathname = pathname === '/' ? '' : pathname;
        copy(`${location.origin}/s/${props.session.checkout}${pathname}`);
        setOpen(false);
    }
    const goToLogs = () => location.href = `/_polo_/session/${props.session.uuid}/logs`;

    return <div className={`helper-session__component background-hover ${open && '--open'}`}>
        <div className={`__visible ${open && '--open'}`} onClick={() => setOpen(open => !open)}>
            <div className="__content">
                <div className="__checkout">
                    <span>On <b className="__checkout-title" title={props.session.checkout}>{props.session.checkout}</b></span>
                </div>
                {maxAgeVisible && <div className="__info">
                    <SessionMaxAge />
                </div>}
            </div>
            <span className="__icon-container">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 -3 24 24" stroke="currentColor" width={16} height={16}>
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
            </span>    
        </div>
        
        <div className="__collapsible">
            {props.session.configuration?.isDefault &&
                <div className="__shortcut" onClick={copyPermalink}>
                    <span className="__desc">Copy permalink</span>
                    <div className="__icon-container">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width={16} height={16}>
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                        </svg>
                    </div>
                </div>
            }
            <div className="__shortcut" onClick={goToLogs}>
                <span className="__desc">View logs</span>
                <div className="__icon-container">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width={16} height={16}>
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16m-7 6h7" />
                    </svg>
                </div>
            </div>
            <div className="__shortcut" onClick={detach}>
                <span className="__desc">Exit</span>
                <div className="__icon-container">
                    <svg className="__icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width={16} height={16}>
                        <path fillRule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 102 0V4a1 1 0 00-1-1zm10.293 9.293a1 1 0 001.414 1.414l3-3a1 1 0 000-1.414l-3-3a1 1 0 10-1.414 1.414L14.586 9H7a1 1 0 100 2h7.586l-1.293 1.293z" clipRule="evenodd" />
                    </svg>
                </div>
            </div>
        </div>
    </div>
}