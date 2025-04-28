# Updated_Thornol_Ros

## Subscribed ROS Topics

As of now, this node subscribes to the following topics (with their default names):

- **odom** (`nav_msgs/msg/Odometry`): Odometry data
- **map** (`nav_msgs/msg/OccupancyGrid`): Map data
- **/scan** (`sensor_msgs/msg/LaserScan`): Laser scan data from the robot's LIDAR
- **plan** (`nav_msgs/msg/Path`): Global path plan
- **local_plan** (`nav_msgs/msg/Path`): Local path plan
- **map_metadata** (`nav_msgs/msg/MapMetaData`): Map metadata
- **navigation_feedback** (`std_msgs/msg/String`): Navigation feedback
- **global_costmap/costmap** (`nav_msgs/msg/OccupancyGrid`): Global costmap
- **soc** (`std_msgs/msg/Float32`): State of charge
- **soh** (`std_msgs/msg/Float32`): State of health
- **amcl_pose** (`geometry_msgs/msg/PoseWithCovarianceStamped`): AMCL pose estimate
- **filtered_cloud** (`sensor_msgs/msg/PointCloud2`): Filtered point cloud data

> Note: Topic names can be customized via the `topic_maps.json` configuration file.
