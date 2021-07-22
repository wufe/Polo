import { APIRequestResult } from '@polo/common/api/common';
import { ISession } from '@polo/common/state/models';
import { SessionStatus } from '@polo/common/state/models/session-model-enums';
import React from 'react';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import duration from 'dayjs/plugin/duration';
dayjs.extend(relativeTime);
dayjs.extend(duration);
import './application-session.scss';
import { observer } from 'mobx-react-lite';
import { useModal } from '@/components/modal/modal-hooks';
import { CommitModal } from '../../../shared/commit-modal';
import { ApplicationSessionModal } from './modal/application-session-modal';
import { ApplicationSessionDeletionModal } from './modal/application-session-deletion-modal';
import { useClipboard } from '@polo/common/components/hooks/use-clipboard';
import { useHistory } from 'react-router-dom';
import loading from '@/assets/loading.svg';
import { Button } from '@polo/common/components/elements/button/button';
import { LoginIcon } from '@polo/common/components/elements/icons/login/login-icon';
import { HorizontalDotsIcon } from '@polo/common/components/elements/icons/horizontal-dots/horizontal-dots-icon';

export const validAgeValue = 1;

export const ApplicationSession = observer((props: { session: ISession }) => {

    const { show, hide } = useModal();
    const copy = useClipboard();
    const history = useHistory();

    const getOptionsModalName = (uuid = props.session.uuid) => `session-${uuid}`;
    const deleteSessionModalName = `${getOptionsModalName()}-session-deletion`;
    const getCommitMessageModalName = (uuid = props.session.uuid) => `${getOptionsModalName(uuid)}-commit`;

    const attachToSession = async () => {
        const track = await props.session.track();
        if (track.result === APIRequestResult.SUCCEEDED) {
            location.href = '/';
        }
    }
    const killSession = async (session: ISession) => {
        hide();
        await session.kill();
    }

    const copySmartURL = () => {
        copy(`${location.origin}${props.session.smartURL}`);
        hide();
    }

    const copyPermalink = () => {
        copy(`${location.origin}${props.session.permalink}`);
        hide();
    }

    const showLogs = (session?: ISession) => {
        if (!session) return;
        hide();
        history.push(`/_polo_/session/${session.uuid}/logs`);
    }

    const openAPIDocument = (session?: ISession) => {
        if (!session) return;
        window.open(`/_polo_/api/session/${session.uuid}`, '_blank');
    }

    const openCommitModal = (session?: ISession) => {
        if (!session) return;
        hide();
        show(getCommitMessageModalName(session.uuid));
    }

    const showLoadingIcon = props.session.status === SessionStatus.STARTING ||
        props.session.beingReplacedBySession;

    const showStartedIcon = props.session.status === SessionStatus.STARTED && !showLoadingIcon;

    return <div
        className="application-session">
        <a className="__content" onClick={attachToSession}>
            <div className="w-6 flex mr-2">
                {showStartedIcon && 
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 20 20"
                        fill={colorByStatus(props.session.status)}
                        className="w-6 h-6 mt-0.5">
                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>}
                {showLoadingIcon && <img src={loading} className="w-4 mt-1 ml-1" />}
            </div>
            <div className="__title-container">
                <span
                    className="__title">
                    {props.session.displayName}
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
                    {props.session.age >= validAgeValue && <span className="__subtitle-item">
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
        </a>
        <span className="text-center whitespace-nowrap flex flex-nowrap items-start">
            <Button
                success
                hideOnMobile
                onClick={attachToSession}
                label="Enter"
                icon={<LoginIcon />} />
            <Button
                ghost
                onClick={() => show(getOptionsModalName())}
                icon={<HorizontalDotsIcon />} />
        </span>
        <ApplicationSessionModal
            session={props.session}
            name={getOptionsModalName()}
            onAPISelect={openAPIDocument}
            onCommitMessageSelect={openCommitModal}
            onEnterSessionSelect={attachToSession}
            onSessionDeletionSelect={() => show(deleteSessionModalName)}
            onCopySmartURLSelect={copySmartURL}
            onCopyPermalinkSelect={copyPermalink}
            onShowLogsSelect={showLogs} />
        <CommitModal
            name={getCommitMessageModalName()}
            title={props.session.displayName}
            commitAuthorEmail={props.session.commitAuthorEmail}
            commitAuthorName={props.session.commitAuthorName}
            commitDate={props.session.commitDate}
            commitMessage={props.session.commitMessage} />
        {props.session.beingReplacedBySession && <CommitModal
            name={getCommitMessageModalName(props.session.beingReplacedBySession.uuid)}
            title={props.session.beingReplacedBySession.displayName}
            commitAuthorEmail={props.session.beingReplacedBySession.commitAuthorEmail}
            commitAuthorName={props.session.beingReplacedBySession.commitAuthorName}
            commitDate={props.session.beingReplacedBySession.commitDate}
            commitMessage={props.session.beingReplacedBySession.commitMessage} />}
        <ApplicationSessionDeletionModal
            name={deleteSessionModalName}
            session={props.session}
            onApplicationDeletionSelected={() => killSession(props.session)} />
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