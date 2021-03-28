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

    const requestData = async () => {
        await props.app.retrieveApplications();
        await props.app.retrieveAllSessions();
    }

    useEffect(() => {
        requestData();
        const interval = setInterval(() => requestData(), 2000);
        return () => clearInterval(interval);
    }, []);

    useEffect(() => {
        setOpenToggleEnabled(props.app.applications.size > 1);
    }, [props.app.applications.size])

    const toggleApplication = (name: string) => () => {
        if (props.app.applications.size > 1)
            setOpenApplications(a => ({ ...a, [name]: !a[name] }));
    }

    return <div className="font-quicksand w-full py-8 pb-12">
        <div className="w-full mx-auto lg:max-w-1500 px-5">
            <div className="flex">
                <div className="py-0 pr-5 hidden lg:block flex-shrink-0 w-3/12">
                <div className="mb-3 text-lg lg:text-xl font-medium text-nord1 dark:text-nord5">Applications</div>
                    {(values(props.app.applications) as any as IApplication[]).map((application, index) =>
                        <div
                            key={index}
                            className="cursor-pointer bg-nord4 dark:bg-nord0 px-5 py-3 rounded-md text-sm lg:text-base">{application.configuration.name}</div>)}
                </div>
                <div className="flex-grow min-w-0">
                    {(values(props.app.applications) as any as IApplication[]).map((application, index) =>
                        <Application
                            isOpen={!!openApplications[application.configuration.name] || !openToggleEnabled}
                            onToggle={toggleApplication(application.configuration.name)}
                            toggleEnabled={openToggleEnabled}
                            key={index}
                            sessions={props.app.sessionsByApplicationName[application.configuration.name]}
                            application={application} />)}
                </div>
            </div>
        </div>
        
    </div>;
})

export default Dashboard;