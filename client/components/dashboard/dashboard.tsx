import React, { useEffect } from 'react';
import { observer } from 'mobx-react-lite';
import { IApp, IService } from '@/state/models';
import { Service } from './service/service';
import './dashboard.scss';
import { values } from 'mobx';

type TProps = {
    app: IApp;
}

export const Dashboard = observer((props: TProps) => {

    const requestData = async () => {
        await props.app.retrieveServices();
        await props.app.retrieveAllSessions();
    }

    useEffect(() => {

        requestData();

        const interval = setInterval(() => requestData(), 2000);
        
        return () => clearInterval(interval);
    }, [])

    return <div className="dashboard__component font-quicksand">
        <h1 className="text-3xl font-light">Dashboard</h1>
        <section>
            <h2 className="pl-3 text-2xl font-light dark:text-gray-300">Services</h2>
            {(values(props.app.services) as any as IService[]).map((service, index) =>
                <Service
                    key={index}
                    sessions={props.app.sessionsByServiceName[service.name]}
                    service={service} />)}
        </section>
    </div>;
})

export default Dashboard;