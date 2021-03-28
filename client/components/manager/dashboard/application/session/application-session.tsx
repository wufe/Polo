import { APIRequestResult } from '@/api/common';
import { ISession } from '@/state/models';
import { SessionStatus } from '@/state/models/session-model-enums';
import React from 'react';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import duration from 'dayjs/plugin/duration';
dayjs.extend(relativeTime);
dayjs.extend(duration);
import './application-session.scss';
import { observer } from 'mobx-react-lite';
import { useModal } from '@/components/manager/modal/modal-hooks';
import { ApplicationSessionCommitModal } from './modal/application-session-commit-modal';
import { ApplicationSessionModal } from './modal/application-session-modal';

export const noExpirationAgeValue = -1;

export const ApplicationSession = observer((props: { session: ISession }) => {

    const { show, hide } = useModal();
    const optionsModalName = `session-${props.session.uuid}`;
    const commitMessageModalName = `${optionsModalName}-commit`;

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

    return <div
        className="application-session">
        <div className="__content" onClick={() => attachToSession(props.session)}>
            <div className="w-6 flex mr-2">
                {props.session.status === SessionStatus.STARTED && <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 20 20"
                    fill={colorByStatus(props.session.status)}
                    className="w-6">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>}
            </div>
            {/* <span className="w-5 text-center inline-block">
                {
                    (
                        props.session.status === SessionStatus.STARTING ||
                        props.session.beingReplaced
                    ) &&
                    <img src={loading} width={12} height={12} className="mr-1" />}
            </span> */}
            {/* <span className="text-xs uppercase font-bold" style={{ color: colorByStatus(props.session.status) }}>{props.session.status}</span> */}
            <div className="__title-container">
                <span
                    className="__title">
                    {props.session.checkout}
                </span>
                <div className="__subtitle-container">
                    <span className="__subtitle-item">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            className="w-3 h-3 mr-1">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                        <span className="whitespace-nowrap">{dayjs(props.session.createdAt).fromNow()}</span>
                    </span>
                    {props.session.age > noExpirationAgeValue && <span className="__subtitle-item">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            className="w-3 h-3 mr-1">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span className="whitespace-nowrap">{dayjs.duration(formatAge(props.session.age)).format('HH:mm:ss')}</span>
                    </span>}
                </div>
            </div>
        </div>

        {/* {props.session.age > noExpirationAgeValue && <span className="text-xs uppercase text-gray-500">
                            <span className="hidden lg:inline-block pr-1">Expires in </span><span>{props.session.age}s</span>
                        </span>} */}
        <span className="text-center whitespace-nowrap flex flex-nowrap items-start">
            <span className="__button --success --hide-on-mobile" onClick={() => attachToSession(props.session)}>
                <span>Enter</span>
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
                </svg>
            </span>
            <span className="__button --ghost inline-flex" onClick={() => show(optionsModalName)}>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M6 10a2 2 0 11-4 0 2 2 0 014 0zM12 10a2 2 0 11-4 0 2 2 0 014 0zM16 12a2 2 0 100-4 2 2 0 000 4z" />
                </svg>
            </span>

            {/* <span className="__button --danger" onClick={() => killSession(session)}>Delete</span> */}
        </span>
        <ApplicationSessionModal
            session={props.session}
            name={optionsModalName}
            onCommitMessageSelect={() => show(commitMessageModalName)}
            onEnterSessionSelect={() => attachToSession(props.session)} />
        <ApplicationSessionCommitModal session={props.session} name={commitMessageModalName} />
    </div>
});

function formatAge(seconds: number): { hours: number; minutes: number; seconds: number } {
    const secondsInHour = 60 * 60;
    const hours = Math.floor(seconds / secondsInHour);
    seconds = seconds - (hours * secondsInHour);
    const minutes = Math.floor(seconds / 60);
    seconds = seconds - (minutes * 60);
    return { hours, minutes, seconds };
}

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