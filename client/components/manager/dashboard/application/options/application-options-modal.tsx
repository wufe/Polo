import React, { useState } from 'react';
import './application-options-modal.scss';
import { DefaultModal } from '@/components/manager/modal/default-modal';
import { observer } from 'mobx-react-lite';
import { IApplication, ISession } from '@/state/models';

type TProps = {
    modalName: string;
    applicationName: string;
    failedSessions: ISession[] | null;
};
export const ApplicationOptionsModal = observer((props: TProps) => {

    const [viewFailingSessions, setViewFailingSession] = useState(false);
    const anyFailedSession = props.failedSessions && props.failedSessions.length > 0;

    return <DefaultModal name={props.modalName}>
        <div className="application-options-modal">
            <div className="__header">
                <span className="inline-block max-w-full text-base lg:text-lg font-bold whitespace-nowrap overflow-hidden overflow-ellipsis">
                    {props.applicationName}
                </span>
            </div>
            {!viewFailingSessions && <div className={`__list ${!anyFailedSession && '--disabled'}`}>
                <div className={`__item ${anyFailedSession && '--danger-icon'}`} onClick={() => anyFailedSession && setViewFailingSession(true)}>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span>View failing sessions</span>
                </div>
            </div>}
            {viewFailingSessions && <div className="__list">
                <div className="__item --command" onClick={() => setViewFailingSession(false)}>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 17l-5-5m0 0l5-5m-5 5h12" />
                    </svg>
                    <span className="font-bold">Go back</span>
                </div>
                <div className="__item --multiple-rows">
                    <div className="__row">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                        <span>Introduced new bug</span>
                    </div>
                    <div className="__row --secondary --indented">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span className="text-gray-400 text-sm">4 hours ago</span>
                    </div>
                </div>
                <div className="__item --multiple-rows">
                    <div className="__row">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                        <span>[MERGED PR: 24] Removed old bug</span>
                    </div>
                    <div className="__row --secondary --indented">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span className="text-gray-400 text-sm">5 hours ago</span>
                    </div>
                </div>
                <div className="__item --multiple-rows">
                    <div className="__row">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                        <span>FEAT-35 Broke build process</span>
                    </div>
                    <div className="__row --secondary --indented">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span className="text-gray-400 text-sm">25 days ago</span>
                    </div>
                </div>
            </div>}
        </div>
    </DefaultModal>
})