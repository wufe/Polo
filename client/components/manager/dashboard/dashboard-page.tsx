import React, { useEffect, useState } from 'react';
import { observer } from 'mobx-react-lite';
import { IApp, IApplication } from '@/state/models';
import { Application } from './application/application';
import { values } from 'mobx';
import { Link } from 'react-router-dom';
import { Modal, ModalPortal } from '../modal/modal-portal';

type TProps = {
    app: IApp;
}
export const Dashboard = observer((props: TProps) => {

    const [openApplications, setOpenApplications] = useState<{[k:string]: boolean}>({});
    const [openToggleEnabled, setOpenToggleEnabled] = useState(false);
    const [selectedAppIndex, setSelectedAppIndex] = useState(-1);

    const selectedApplicationLocalStorageKey = 'selected-application-name';

    const requestData = async () => {
        await props.app.retrieveApplications();
        await props.app.retrieveAllSessions();
    }

    useEffect(() => {
        const apps = values(props.app.applications) as any as IApplication[];
        if (selectedAppIndex > -1 || apps.length === 0) return;
        const applicationName = localStorage.getItem(selectedApplicationLocalStorageKey);
        const foundIndex = apps
            .findIndex(app => app.configuration.name === applicationName);
        if (foundIndex > -1) {
            setSelectedAppIndex(foundIndex);
        } else {
            setSelectedAppIndex(0);
        }
    }, [values(props.app.applications)]);

    useEffect(() => {
        requestData();
        const interval = setInterval(() => requestData(), 2000);
        return () => clearInterval(interval);
    }, []);

    useEffect(() => {
        setOpenToggleEnabled(props.app.applications.size > 1);
    }, [props.app.applications.size]);

    const openApplication = (name: string, index: number) => () => {
        setSelectedAppIndex(index);
        localStorage.setItem(selectedApplicationLocalStorageKey, name);
    }

    const selected: IApplication = values(props.app.applications)[selectedAppIndex] as any;

    return <div className="font-quicksand w-full py-8 pb-12">
        <div className="w-full mx-auto lg:max-w-1500 px-5">
            <div className="flex">
                <div className="py-0 pr-5 hidden lg:block flex-shrink-0 w-3/12">
                <div className="mb-3 text-lg lg:text-xl font-medium text-nord1 dark:text-nord5">Applications</div>
                    {(values(props.app.applications) as any as IApplication[]).map((application, index) =>
                        <div
                            key={index}
                            className={`cursor-pointer px-5 py-3 rounded-md text-sm lg:text-base mb-3
                            ${selectedAppIndex === index ? 'bg-nord4 dark:bg-nord0' : ''}`}
                            onClick={openApplication(application.configuration.name, index)}>{application.configuration.name}</div>)}
                </div>
                {!!selected && <div className="flex-grow min-w-0">
                    <Application
                        sessions={props.app.sessionsByApplicationName[selected.configuration.name]}
                        application={selected} />
                </div>}
            </div>
        </div>
        
    </div>;
})

export default Dashboard;